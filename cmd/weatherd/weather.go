package main

import (
	"burlo/config"
	"burlo/pkg/models/weather"
	"burlo/pkg/mqtt"
	"burlo/pkg/openmateo"
	"burlo/pkg/weathergcca"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.LoadV2(*configPath)
	mqttc := mqtt.NewClient(mqtt.Opts{
		Context:     ctx,
		Address:     cfg.Mqtt.Address,
		User:        cfg.Mqtt.User,
		Pass:        []byte(cfg.Mqtt.Pass),
		TopicPrefix: "burlo",
		ClientID:    "weatherd",
	})

	fmt.Println("started")
	defer fmt.Println("stopped")

	var wService weather.WeatherService
	wService, err := openmateo.New(cfg.Location.Latitude, cfg.Location.Longitude)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Poll current conditions every 15 minutes
	// and publish to mqtt
	go func() {
		for {
			current, err := wService.CurrentConditions()
			if err == nil {
				mqttc.Publish(true, "weather/current", current)
			} else {
				mqttc.Publish(false, "error/weather/current", err.Error())
				fmt.Println("[Error] fetching current conditions:", err)
			}
			time.Sleep(15 * time.Minute)
		}
	}()

	// Poll forecast data once per hour
	// and publish to mqtt
	go func() {
		for {
			forecast, err := wService.Forecast24h()
			if err == nil {
				mqttc.Publish(true, "weather/forecast", forecast)
			} else {
				mqttc.Publish(false, "error/weather/forecast", err.Error())
				fmt.Println("[Error] fetching forecast:", err)
			}
			time.Sleep(time.Hour)
		}
	}()

	// Poll air quality health index (AQHI) current+forecast once per hour
	// and publish to mqtt
	go func() {
		for {
			// TODO: location needs to be configurable
			fmt.Println("getting aqhi")
			forecast, err := weathergcca.GetAqhiForecast(3, "Ottawa")
			if err == nil {
				mqttc.Publish(true, "weather/aqhi", forecast)
			} else {
				mqttc.Publish(false, "error/weather/aqhi", err.Error())
				fmt.Println("[Error] fetching aqhi:", err)
			}
			time.Sleep(time.Hour)
		}
	}()

	<-ctx.Done()
}
