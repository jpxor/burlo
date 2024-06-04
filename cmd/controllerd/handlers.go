package main

import (
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
)

var inputs CtrlInput
var inputMutex sync.Mutex
var thermostats = make(map[string]controller.Thermostat)

func onThermostatUpdate(payload []byte) {
	var tstat controller.Thermostat
	err := json.Unmarshal(payload, &tstat)
	if err != nil {
		fmt.Println("onThermostatUpdate:", err)
		return
	}
	inputMutex.Lock()
	defer inputMutex.Unlock()

	// update thermostats mapping
	thermostats[tstat.ID] = tstat

	// find max and mean values
	var maxTemp float32 = 0.0
	var maxDewpoint float32 = 0
	var meanHeatSetpointErr float32 = 0.0
	var meanCoolSetpointErr float32 = 0.0

	for _, tstat := range thermostats {
		maxTemp = max(maxTemp, tstat.Temperature)
		maxDewpoint = max(maxDewpoint, tstat.Dewpoint)
		meanHeatSetpointErr += tstat.Temperature - tstat.HeatSetpoint
		meanCoolSetpointErr += tstat.Temperature - tstat.CoolSetpoint
	}
	meanHeatSetpointErr /= float32(len(thermostats))
	meanCoolSetpointErr /= float32(len(thermostats))

	// update inputs and trigger controller routine
	inputs.Indoor.Temperature = maxTemp
	inputs.Indoor.Dewpoint = maxDewpoint
	inputs.Indoor.HeatSetpointErr = meanHeatSetpointErr
	inputs.Indoor.CoolSetpointErr = meanCoolSetpointErr
	inputs.Ready |= IndoorReady

	tryRunController(inputs)
}

func onForecastUpdate(payload []byte) {
	var data weather.Forecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onForecastUpdate:", err)
		return
	}
	inputMutex.Lock()
	defer inputMutex.Unlock()

	mean := func(list []float32) float32 {
		var sum float32
		for _, f := range list {
			sum += f
		}
		return sum / float32(len(list))
	}
	inputs.Outdoor.T24hHigh = slices.Max(data.Temperature)
	inputs.Outdoor.T24hLow = slices.Min(data.Temperature)
	inputs.Outdoor.T24hMean = mean(data.Temperature)
	inputs.Ready |= ForecastReady

	tryRunController(inputs)
}

func onCurrentWeatherUpdate(payload []byte) {
	var data weather.Current
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onCurrentWeatherUpdate:", err)
		return
	}
	inputMutex.Lock()
	defer inputMutex.Unlock()

	inputs.Outdoor.Temperature = data.Temperature
	inputs.Outdoor.Dewpoint = calculate_dewpoint_simple(data.Temperature, data.RelHumidity)
	inputs.Ready |= CurrentReady

	tryRunController(inputs)
}

// a simple approximation, should err on the side of
// being too high, never too low. Temperature must be
// in celcius. Accureate to within 1 degC when RelH > 50%
func calculate_dewpoint_simple(temp, relH float32) float32 {
	if relH >= 50 && temp >= 25 {
		return temp - ((100 - relH) / 5)
	} else {
		return temp - ((100 - relH) / 4)
	}
}
