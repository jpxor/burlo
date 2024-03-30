package main

import (
	services "burlo/services/model"
	. "burlo/services/protocols/controller"
	"log"
	"sync"
)

type global_vars struct {
	waitgroup sync.WaitGroup
	mutex     sync.Mutex
	Conditions
	Controls
	thermostats map[string]services.Thermostat
}

var global = global_vars{
	thermostats: make(map[string]services.Thermostat),
}

func main() {
	log.Println("Running controller service")

	global.waitgroup.Add(1)
	go controller_update_service()

	global.waitgroup.Add(1)
	go controller_http_server()

	global.waitgroup.Wait()
}
