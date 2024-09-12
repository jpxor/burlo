package main

import (
	"burlo/config"
	"burlo/pkg/mqtt"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	cfg := config.LoadV2(*configPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("started")
	defer fmt.Println("stopped")

	go httpserver(ctx, cfg)

	mqtt.NewClient(mqtt.Opts{
		Context:     ctx,
		Address:     cfg.Mqtt.Address,
		User:        cfg.Mqtt.User,
		Pass:        []byte(cfg.Mqtt.Pass),
		ClientID:    "dashboard",
		TopicPrefix: "burlo",
		Topics: []string{
			"controller/thermostats/#",
			"controller/humidistat/#",
			"weather/current",
			"weather/forecast",
			"weather/aqhi",
		},
		OnPublishRecv: func(topic string, payload []byte) {
			topic = strings.TrimPrefix(topic, "burlo/")

			switch {
			case strings.HasPrefix(topic, "controller/thermostats/"):
				onThermostatUpdate(payload)

			case strings.HasPrefix(topic, "controller/humidistat/"):
				onThermostatUpdate(payload)

			case strings.HasPrefix(topic, "weather/current"):
				onCurrentWeatherUpdate(payload)

			case strings.HasPrefix(topic, "weather/forecast"):
				onForecastUpdate(payload)

			case strings.HasPrefix(topic, "weather/aqhi"):
				onAQHIUpdate(payload)

			default:
				fmt.Println("unhandled topic:", topic)
			}
		},
	})

	// waits for signal
	<-ctx.Done()
}
