package main

import (
	"burlo/config"
	weather "burlo/services/weather/model"
	"burlo/services/weather/openmateo"
	"context"
	"log"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

// WeatherService defines the interface for the backend weather service.
type WeatherService interface {
	CurrentConditions() (weather.Conditions, error)
	TemperatureForcast24h() (weather.Forcast, error)
}

func main() {
	cfg := config.Load("../config/config.toml")

	log.Println("[weather] started weather service")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize the Open-Meteo weather service
	wService, err := openmateo.New(cfg.Weather.Latitude, cfg.Weather.Longitude)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Poll current conditions every 15 minutes
	go func() {
		for {
			current, err := wService.CurrentConditions()
			if err != nil {
				log.Printf("Error fetching CurrentConditions: %v", err)
			} else {
				log.Printf("Current temperature: %.2f°C\r\n", current.Temperature)
				log.Printf("%+v\r\n", current)
			}
			time.Sleep(15 * time.Minute)
		}
	}()

	// Poll forecast data once per hour
	go func() {
		for {
			tforcast, err := wService.TemperatureForcast24h()
			if err != nil {
				log.Printf("Error fetching forecast data: %v", err)
			} else {
				tmax := slices.Max(tforcast.Temperatures)
				tmin := slices.Min(tforcast.Temperatures)
				tavg := Mean(tforcast.Temperatures)
				log.Printf("%+v\r\n", tforcast)
				log.Printf("tmax %.2f°C, tmin %.2f°C, tavg %.2f°C\r\n",
					tmax, tmin, tavg)
			}
			time.Sleep(time.Hour)
		}
	}()

	<-ctx.Done()
	log.Println("[weather] stopped weather service")
}

func Mean(vals []float32) float32 {
	var sum float32
	for _, v := range vals {
		sum += v
	}
	return sum / float32(len(vals))
}
