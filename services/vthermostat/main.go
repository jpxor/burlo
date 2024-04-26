package main

import (
	"burlo/pkg/lockbox"
	"fmt"
	"log"
	"os"
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
	if len(os.Args) < 3 {
		fmt.Println("1. provide path to config file")
		fmt.Println("2. provide path to www directory")
		os.Exit(1)
	}
	configPath := os.Args[1]
	wwwPath := os.Args[2]
	load_controller_addr(configPath)

	log.Println("Running virtual thermostat service")

	global.waitgroup.Add(1)
	go process_thermostat_updates()

	global.waitgroup.Add(1)
	go process_mqtt_updates()

	global.waitgroup.Add(1)
	go go_gadget_web_app(wwwPath)

	// waits for all service type goroutines to complete
	global.waitgroup.Wait()
}
