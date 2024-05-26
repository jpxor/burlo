package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type Opts struct {
	Context       context.Context
	Address       string
	User          string
	Pass          []byte
	ClientID      string
	TopicPrefix   string
	Topics        []string
	OnPublishRecv func(topic string, payload []byte)
}

type Client struct {
	opts Opts
	cm   *autopaho.ConnectionManager
}

func NewClient(opts Opts) *Client {
	client := &Client{opts: opts}

	u, err := url.Parse(fmt.Sprintf("mqtt://%s", opts.Address))
	if err != nil {
		panic(err)
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{u},
		KeepAlive:             20,
		SessionExpiryInterval: 60,
		ConnectUsername:       opts.User,
		ConnectPassword:       opts.Pass,
		OnConnectError: func(err error) {
			fmt.Println("[Error] on mqtt connect:", err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: opts.ClientID,
			OnClientError: func(err error) {
				fmt.Println("[Error] mqtt client error:", err)
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				fmt.Print("mqtt disconnected: ")
				if d.Properties != nil {
					fmt.Println(d.Properties.ReasonString)
				} else {
					fmt.Println(d.ReasonCode)
				}
			},
		},
	}

	var subs []paho.SubscribeOptions
	for _, topic := range opts.Topics {
		subs = append(subs, paho.SubscribeOptions{
			Topic: path.Join(opts.TopicPrefix, topic),
			QoS:   1,
		})
	}

	cliCfg.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		fmt.Println("connected to", u.String())
		if len(subs) > 0 {
			_, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: subs,
			})
			if err != nil {
				fmt.Println("[Error] failed to subscribe:", err)
			}
		}
	}

	if opts.OnPublishRecv != nil {
		cliCfg.ClientConfig.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
			func(pr paho.PublishReceived) (bool, error) {
				opts.OnPublishRecv(pr.Packet.Topic, pr.Packet.Payload)
				return true, nil
			},
		}
	}

	client.cm, err = autopaho.NewConnection(opts.Context, cliCfg)
	if err != nil {
		panic(err)
	}

	if err = client.cm.AwaitConnection(opts.Context); err != nil {
		panic(err)
	}

	return client
}

func (c *Client) Publish(retain bool, topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.cm.Publish(c.opts.Context, &paho.Publish{
		QoS:     1,
		Topic:   path.Join(c.opts.TopicPrefix, topic),
		Retain:  retain,
		Payload: payload,
	})
	return err
}
