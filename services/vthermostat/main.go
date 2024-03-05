package main

import (
	"log"
	"sync"
	"vthermostat/lockbox"
)

// tracks current state of all sensors and outdoor conditions
//  - listens on mqtt topics [done]
//  - state:
//     - sensor name, temperature, humidity, dewpoint, last update
//     - outdoor temperature, humidity, dewpoint, windspeed
//     - average outdoor temperature over 24hr
//  - web interface for status

type global_vars struct {
	thermostats  *lockbox.LockBox[map[string]Thermostat]
	history      *lockbox.LockBox[[]HistoryData]
	notify_queue chan Thermostat
	waitgroup    sync.WaitGroup
}

var global = global_vars{
	thermostats:  lockbox.New(map[string]Thermostat{}),
	history:      lockbox.New([]HistoryData{}),
	notify_queue: make(chan Thermostat),
}

func main() {
	log.Println("Running virtual thermostat service")

	global.waitgroup.Add(1)
	go process_thermostat_updates()

	global.waitgroup.Add(1)
	go process_mqtt_updates()

	// waits for all service type goroutines to complete
	global.waitgroup.Wait()
}
