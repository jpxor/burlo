package main

import (
	"burlo/config"
	"burlo/pkg/lockbox"
	"flag"
	"fmt"
	"sync"
)

// tracks current state of all sensors and outdoor conditions
//  - listens on mqtt topics [done]
//  - state:
//     - sensor name, temperature, humidity, dewpoint, last update
//     - outdoor temperature, humidity, dewpoint, windspeed
//     - average outdoor temperature over 24hr
//  - web interface for status

type global_vars struct {
	thermostats *lockbox.LockBox[map[string]Thermostat]
	history     *lockbox.LockBox[[]HistoryData]
	waitgroup   sync.WaitGroup
}

var global = global_vars{
	thermostats: lockbox.New(map[string]Thermostat{}),
	history:     lockbox.New([]HistoryData{}),
}

func main() {
	configPath := flag.String("c", "", "Path to config file")
	wwwPath := flag.String("w", "", "Path to thermostatd webserver root")
	flag.Parse()

	cfg := config.LoadV2(*configPath)
	load_controller_addr(cfg)

	fmt.Println("started")
	defer fmt.Println("stopped")

	global.waitgroup.Add(1)
	go process_thermostat_updates(cfg)

	global.waitgroup.Add(1)
	go process_mqtt_updates(cfg)

	global.waitgroup.Add(1)
	go go_gadget_web_app(*wwwPath)

	// waits for all service type goroutines to complete
	global.waitgroup.Wait()
}
