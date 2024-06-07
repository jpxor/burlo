package main

import (
	"burlo/config"
	"burlo/pkg/mqtt"
	"burlo/pkg/ntfy"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var notify ntfy.Notify

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	cfg := config.LoadV2(*configPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("started")
	defer fmt.Println("stopped")

	notify = ntfy.New(cfg.ServiceHTTPAddresses.NtfyServer, "burlo")

	initPhidgetsClient(cfg.ServiceHTTPAddresses.Actuators)
	go httpserver(ctx, cfg)

	mqtt.NewClient(mqtt.Opts{
		Context:     ctx,
		Address:     cfg.Mqtt.Address,
		User:        cfg.Mqtt.User,
		Pass:        []byte(cfg.Mqtt.Pass),
		ClientID:    "controllerd",
		TopicPrefix: "burlo",
		Topics: []string{
			"controller/thermostats/#",
			"controller/humidistat/#",
			"weather/current",
			"weather/forecast",
		},
		OnPublishRecv: func(topic string, payload []byte) {
			topic = strings.TrimPrefix(topic, "burlo/")

			switch true {
			case strings.HasPrefix(topic, "controller/thermostats/"):
				onThermostatUpdate(payload)

			case strings.HasPrefix(topic, "controller/humidistat/"):
				onThermostatUpdate(payload)

			case strings.HasPrefix(topic, "weather/current"):
				onCurrentWeatherUpdate(payload)

			case strings.HasPrefix(topic, "weather/forecast"):
				onForecastUpdate(payload)

			default:
				fmt.Println("unhandled topic:", topic)
			}
		},
	})

	// waits for signal
	<-ctx.Done()
}
