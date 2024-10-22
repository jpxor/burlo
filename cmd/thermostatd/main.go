package main

import (
	"burlo/config"
	"burlo/pkg/models/controller"
	"burlo/pkg/mqtt"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// thermostatd: collects temperature and humidity data from various sensors,
// computes the dewpoint, and then writes the sensor data in a common format
// to the controller mqtt topic. This service also allows setting thermostat
// heat and cool setpoints, and change its name

var publisher *mqtt.Client

var mutex sync.Mutex
var thermostats = make(map[string]controller.Thermostat)

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	cfg := config.LoadV2(*configPath)

	fmt.Println("started")
	defer fmt.Println("stopped")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisher = mqtt.NewClient(mqtt.Opts{
		Context:     ctx,
		Address:     cfg.Mqtt.Address,
		User:        cfg.Mqtt.User,
		Pass:        []byte(cfg.Mqtt.Pass),
		ClientID:    "thermostatd",
		TopicPrefix: "burlo",
	})

	go monitor_sonoff_zigbee2mqtt(ctx, cfg)
	go monitor_picosense_mqtt(ctx, cfg)
	go http_server(ctx, cfg)

	// waits for signal
	<-ctx.Done()
}
