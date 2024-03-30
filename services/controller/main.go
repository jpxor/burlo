package main

import (
	"log"
	"sync"
)

type global_vars struct {
	waitgroup sync.WaitGroup
	mutex     sync.Mutex
	Conditions
	Controls
	thermostats map[string]Thermostat
}

var global = global_vars{
	thermostats: make(map[string]Thermostat),
}

func main() {
	log.Println("[controller_update_service] started")
	defer log.Println("[controller_update_service] stopped")

	initControls()

	global.waitgroup.Add(1)
	go controller_http_server()

	global.waitgroup.Wait()
}

func initControls() {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	global.Controls = Controls{
		Circulator: ControlMode[Mode]{
			Mode:       OFF,
			ValidModes: []Mode{ON, OFF},
		},
		Heatpump: ControlMode[Mode]{
			Mode:       HEAT,
			ValidModes: []Mode{HEAT, COOL},
		},
		SupplyTemp: ControlValue[float32]{
			Value: design_indoor_air_temperature,
			Min:   min_cooling_supply_temperature,
			Max:   max_supply_temperature,
		},
	}
}
