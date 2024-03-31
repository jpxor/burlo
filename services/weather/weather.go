package main

import (
	"burlo/config"
	"burlo/pkg/lockbox"
	protocol "burlo/services/protocols"
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

var outdoor_conditions = lockbox.New(protocol.OutdoorConditions{})
var lastForcastUpdate time.Time

func main() {
	log.Println("[weather] started weather service")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load("../config/config.toml")
	initHttpClient(cfg.Services.ControllerAddr)

	// Initialize the Open-Meteo weather service
	wService, err := openmateo.New(cfg.Weather.Latitude, cfg.Weather.Longitude)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Poll current conditions every 15 minutes
	// and send weather data to controller
	go func() {
		for {
			current, err := wService.CurrentConditions()
			if err != nil {
				log.Printf("Error fetching CurrentConditions: %v", err)
			} else {
				log.Printf("[weather] %+v\r\n", current)

				// wait to have recent forcast data before sending
				// weather data to the controller
				for time.Since(lastForcastUpdate) > time.Hour {
					time.Sleep(time.Second)
				}
				conditions, lbk := outdoor_conditions.Take()
				conditions.OutdoorAirTemp = current.Temperature

				notify_controller(conditions)
				outdoor_conditions.Put(conditions, lbk)
			}
			time.Sleep(15 * time.Minute)
		}
	}()

	// Poll forecast data once per hour
	go func() {
		var Mean = func(vals []float32) float32 {
			var sum float32
			for _, v := range vals {
				sum += v
			}
			return sum / float32(len(vals))
		}
		for {
			tforcast, err := wService.TemperatureForcast24h()
			if err != nil {
				log.Printf("Error fetching forecast data: %v", err)
			} else {
				tmax := slices.Max(tforcast.Temperatures)
				tmin := slices.Min(tforcast.Temperatures)
				tavg := Mean(tforcast.Temperatures)
				log.Printf("[weather] tmax %.2f°C, tmin %.2f°C, tavg %.2f°C\r\n",
					tmax, tmin, tavg)

				conditions, lbk := outdoor_conditions.Take()
				conditions.OutdoorAir24hLow = tmin
				conditions.OutdoorAir24hAvg = tavg
				conditions.OutdoorAir24hHigh = tmax

				lastForcastUpdate = time.Now()
				outdoor_conditions.Put(conditions, lbk)
			}
			time.Sleep(time.Hour)
		}
	}()

	<-ctx.Done()
	log.Println("[weather] stopped weather service")
}
