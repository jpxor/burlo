package main

import (
	. "burlo/services/controller/model"
	services "burlo/services/model"
	"log"
	"sync"
	"time"
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

func update_outdoor_conditions(odc OutdoorConditions) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	global.conditions.OutdoorConditions = odc
	global.conditions.OutdoorConditions.LastUpdate = time.Now()
	global.state = system_update(global.state, global.conditions)
}

func update_indoor_conditions(tstat services.Thermostat) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	// update thermostat cache, recalculate
	// average setpoint error
	global.thermostats[tstat.ID] = tstat

	var idc IndoorConditions
	for _, tstat := range global.thermostats {
		idc.IndoorAirTempMax = max(idc.IndoorAirTempMax, tstat.State.Temperature)
		idc.DewPoint = max(idc.DewPoint, tstat.State.DewPoint)
		switch global.state.Heatpump.Mode.Value {
		case HEAT:
			idc.SetpointError += tstat.State.Temperature - tstat.HeatSetpoint
		case COOL:
			idc.SetpointError += tstat.State.Temperature - tstat.CoolSetpoint
		}
	}
	idc.SetpointError /= float32(len(global.thermostats))

	global.conditions.IndoorConditions = idc
	global.conditions.IndoorConditions.LastUpdate = time.Now()
	global.state = system_update(global.state, global.conditions)
}
