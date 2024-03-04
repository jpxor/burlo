package main

import (
	"log"
	"sync"
)

// tracks current state of all sensors and outdoor conditions
//  - listens on mqtt topics
//  - state:
//     - sensor name, temperature, humidity, dewpoint, last update
//     - outdoor temperature, humidity, dewpoint, windspeed
//     - average outdoor temperature over 24hr
//  - web interface for status

type global_vars struct {
	thermostats *RWMap[string, *Thermostat]
	waitgroup   sync.WaitGroup
}

var global = global_vars{
	thermostats: NewRWMap[string, *Thermostat](),
}

var notify_controller = func(tstat Thermostat) {
	log.Fatalln("notify_controller is not yet initialized")
}

func main() {
	log.Println("Running virtual thermostat service")

	global.waitgroup.Add(1)
	go run_mqtt_sensors_client()

	// waits for all service type goroutines to complete
	global.waitgroup.Wait()
}
