package main

import (
	protocol "burlo/services/protocols"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func process_mqtt_updates() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	clientID := "virt-thermostat"
	host := "192.168.50.193"
	port := "1883"
	user := "hvac"
	pass := "hvac_pass"
	topic := "zigbee2mqtt/thermostats/#"
	qos := byte(1) // at least once

	u, err := url.Parse(fmt.Sprintf("mqtt://%s:%s", host, port))
	if err != nil {
		panic(err)
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{u},
		KeepAlive:             20,
		SessionExpiryInterval: 60,
		ConnectUsername:       user,
		ConnectPassword:       []byte(pass),
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Printf("[mqtt] connected to %s\r\n", u.String())
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: topic, QoS: qos},
				},
			}); err != nil {
				log.Fatalf("[mqtt] failed to subscribe (%s)", err)
			}
			log.Printf("[mqtt] subscribed to %s\r\n", topic)
		},
		OnConnectError: func(err error) {
			log.Printf("[mqtt] error whilst attempting connection: %s\n", err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					return mqtt_message_handler(pr.Packet.Topic, pr.Packet.Payload), nil
				}},
			OnClientError: func(err error) {
				log.Printf("[mqtt] client error: %s\n", err)
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("[mqtt] server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Printf("[mqtt] server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	c, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		log.Fatalln(err)
	}

	if err = c.AwaitConnection(ctx); err != nil {
		log.Fatalln(err)
	}

	<-ctx.Done() // waits for interrupt signal
	log.Println("[mqtt] stopping")

	<-c.Done() // waits for clean shutdown of autopaho mqtt connection
	log.Println("[mqtt] client cleanup complete")

	global.waitgroup.Done()
}

func mqtt_message_handler(topic string, payload []byte) bool {
	prefix := "zigbee2mqtt/thermostats/"
	if !strings.HasPrefix(topic, prefix) {
		return false
	}
	name := strings.TrimPrefix(topic, prefix)
	id, _, _ := strings.Cut(name, "/")

	// id needs to be safe to use in URL paths as well
	// as in css class names
	id = url.PathEscape(id)
	id = strings.ReplaceAll(id, "%", "_")

	var new_state protocol.SensorData
	err := json.Unmarshal(payload, &new_state)
	if err != nil {
		log.Printf("[mqtt] failed to parse payload: %s --> %s\r\n", id, string(payload))
		return true
	}

	thermostats, lbk := global.thermostats.Take()
	tstat, found := thermostats[id]
	if !found {
		// new sensor detected, need to create a new
		// thermostat setpoint contoller to go with it
		tstat = Thermostat{
			ID:           id,
			Name:         name,
			HeatSetpoint: 20, // default
			CoolSetpoint: 24, // default
		}
		log.Printf("[mqtt] new thermostat %s\r\n", id)
	}
	tstat.From(new_state)

	log.Printf("[mqtt] %s --> %s\r\n", id, string(payload))
	go async_process_thermostat_update(tstat)

	thermostats[id] = tstat
	global.thermostats.Put(thermostats, lbk)
	return true
}
