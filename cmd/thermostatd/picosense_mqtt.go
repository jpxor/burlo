package main

import (
	"burlo/config"
	"burlo/pkg/models/controller"
	"burlo/pkg/mqtt"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type PicosenseSensor struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

func monitor_picosense_mqtt(ctx context.Context, cfg config.ServiceConf) {
	mqtt.NewClient(mqtt.Opts{
		Context:  ctx,
		Address:  cfg.Mqtt.Address,
		User:     cfg.Mqtt.User,
		Pass:     []byte(cfg.Mqtt.Pass),
		ClientID: "thermostatd_picosense",
		Topics: []string{
			"picosense/#",
		},
		OnPublishRecv: func(topic string, payload []byte) {
			topic = strings.TrimPrefix(topic, "picosense/")

			// expected format: "id/name"
			id, name, ok := strings.Cut(topic, "/")
			if !ok {
				name = id
			}

			var sensor PicosenseSensor
			err := json.Unmarshal(payload, &sensor)
			if err != nil {
				fmt.Printf("failed to parse mqtt payload: %s\r\n", string(payload))
				return
			}

			go update(controller.Thermostat{
				ID:           id,
				Name:         name,
				DewpointOnly: false,
				Temperature:  sensor.Temperature,
				Humidity:     sensor.Humidity,
				Battery:      100,
				LinkQuality:  100, // TODO: get wlan.status('rssi') ?
			})
		},
	})

	// waits for signal
	<-ctx.Done()
}
