package main

import (
	"burlo/config"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

// configs
var cooling_enabled bool = true
var min_cooling_supply_temperature float32 = 12
var comfort_cooling_supply_temperature float32 = 18
var max_supply_temperature float32 = 40.55
var design_supply_temperature float32 = 40.55
var design_outdoor_air_temperature float32 = -25
var design_indoor_air_temperature float32 = 20
var zero_load_outdoor_air_temperature float32 = 16
var cooling_mode_high_temp_trigger float32 = 28

func loadConfig(path string) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	cfg := config.Load(path)
	cooling_enabled = cfg.Controller.Cooling.Enabled
	min_cooling_supply_temperature = cfg.Controller.Cooling.OvernightBoostTemperature
	comfort_cooling_supply_temperature = cfg.Controller.Cooling.CoolingSupplyTemperature
	cooling_mode_high_temp_trigger = cfg.Controller.Cooling.CoolingTriggerTemperature
	max_supply_temperature = cfg.Controller.Heating.MaxSupplyTemperature
	design_supply_temperature = cfg.Controller.Heating.DesignLoadSupplyTemperature
	design_outdoor_air_temperature = cfg.Controller.Heating.DesignLoadOutdoorAirTemperature
	design_indoor_air_temperature = cfg.Controller.Heating.ZeroLoadSupplyTemperature
	zero_load_outdoor_air_temperature = cfg.Controller.Heating.ZeroLoadOutdoorAirTemperature

	update_controls_locked()
}

func controller_config_watcher(path string) {
	defer global.waitgroup.Done()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[controller_config_watcher] started")
	defer log.Println("[controller_config_watcher] stopped")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file modified:", event.Name)
				loadConfig(path)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("[controller_config_watcher] watcher error:", err)

		case <-ctx.Done():
			return
		}
	}
}
