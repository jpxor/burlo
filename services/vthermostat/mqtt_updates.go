package main

import (
	"burlo/config"
	"burlo/pkg/mqtt"
	protocol "burlo/services/protocols"
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func process_mqtt_updates(cfg config.ServiceConf) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mqtt.NewClient(mqtt.Opts{
		Context:       ctx,
		Address:       cfg.Mqtt.Address,
		User:          cfg.Mqtt.User,
		Pass:          []byte(cfg.Mqtt.Pass),
		ClientID:      "thermostatd",
		Topics:        []string{"zigbee2mqtt/thermostats/#"},
		OnPublishRecv: mqtt_message_handler,
	})

	<-ctx.Done() // waits for interrupt signal
	log.Println("[mqtt] stopping")

	global.waitgroup.Done()
}

func mqtt_message_handler(topic string, payload []byte) {
	prefix := "zigbee2mqtt/thermostats/"
	if !strings.HasPrefix(topic, prefix) {
		return
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
		return
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
}
