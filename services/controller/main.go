package main

import (
	. "burlo/services/controller/model"
	services "burlo/services/model"
	"log"
	"sync"
)

type global_vars struct {
	waitgroup   sync.WaitGroup
	mutex       sync.Mutex
	state       SystemStateV2
	conditions  ControlConditions
	thermostats map[string]services.Thermostat
}

var global = global_vars{
	state: SystemStateV2{
		Circulator{
			initValue(false),
		},
		Heatpump{
			Mode:          initValue(HEAT),
			TsTemperature: initValue(float32(20)),
			TsCorrection:  initValue(float32(0)),
		},
	},
	conditions: ControlConditions{
		IndoorConditions{},
		OutdoorConditions{},
	},
	thermostats: make(map[string]services.Thermostat),
}

func main() {
	log.Println("Running controller service")

	global.waitgroup.Add(1)
	go controller_http_server()

	global.waitgroup.Wait()
}
