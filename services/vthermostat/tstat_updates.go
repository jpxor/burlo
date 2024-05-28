package main

import (
	"burlo/config"
	"burlo/pkg/mqtt"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var tstat_update_queue = make(chan Thermostat)

// this is called from the mqtt message handler callback.
// that handler can't block so it spawns a goroutine to handle
// the sensor updates, and that goroutine is queued up here
// to ensure only one at a time are processed (no need to parallel)
func async_process_thermostat_update(tstat Thermostat) {
	tstat_update_queue <- tstat
}

// loop forever pulling tstats from the channel
func process_thermostat_updates(cfg config.ServiceConf) {
	defer global.waitgroup.Done()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mqttc := mqtt.NewClient(mqtt.Opts{
		Context:     ctx,
		Address:     cfg.Mqtt.Address,
		User:        cfg.Mqtt.User,
		Pass:        []byte(cfg.Mqtt.Pass),
		ClientID:    "thermostatd",
		TopicPrefix: "burlo",
	})
	retain := true

	log.Println("[process_thermostat_updates] started")
	for {
		select {
		case tstat := <-tstat_update_queue:
			notify_controller(tstat)
			mqttc.Publish(retain, fmt.Sprintf("controller/thermostats/%s", tstat.ID), tstat)
			update_history(tstat)

		case <-ctx.Done():
			log.Println("[process_thermostat_updates] stopped")
			return
		}
	}
}
