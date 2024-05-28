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

// SNZB-02D
type SonoffSensor struct {
	Battery     int32   `json:"battery"`
	LinkQuality int32   `json:"linkquality"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
}

func monitor_sonoff_zigbee2mqtt(ctx context.Context, cfg config.ServiceConf) {
	mqtt.NewClient(mqtt.Opts{
		Context:  ctx,
		Address:  cfg.Mqtt.Address,
		User:     cfg.Mqtt.User,
		Pass:     []byte(cfg.Mqtt.Pass),
		ClientID: "thermostatd_zigbee2mqtt",
		Topics:   []string{"zigbee2mqtt/thermostats/#"},
		OnPublishRecv: func(topic string, payload []byte) {

			// expected format: "name/id"
			name := strings.TrimPrefix(topic, "zigbee2mqtt/thermostats/")
			id, _, _ := strings.Cut(name, "/")

			var sensor SonoffSensor
			err := json.Unmarshal(payload, &sensor)
			if err != nil {
				fmt.Printf("failed to parse mqtt payload: %s\r\n", string(payload))
				return
			}

			go update(controller.Thermostat{
				ID:          id,
				Name:        name,
				Temperature: sensor.Temperature,
				Humidity:    sensor.Humidity,
				Battery:     sensor.Battery,
				LinkQuality: sensor.LinkQuality,
			})
		},
	})

	// waits for signal
	<-ctx.Done()
}
