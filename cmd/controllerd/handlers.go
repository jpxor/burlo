package main

import (
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"burlo/pkg/weathergcca"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"
)

var inputs = CtrlInput{
	ModeOverride:  DX2W_AUTO,
	StateOverride: DX2W_STATE_AUTO,
}
var inputMutex sync.Mutex
var thermostats = make(map[string]controller.Thermostat)
var humidistats = make(map[string]controller.Thermostat)

func onThermostatUpdate(payload []byte) {
	var tstat controller.Thermostat
	err := json.Unmarshal(payload, &tstat)
	if err != nil {
		fmt.Println("onThermostatUpdate:", err)
		return
	}
	inputMutex.Lock()
	defer inputMutex.Unlock()

	if tstat.Battery < 20 {
		notify.Publish("sensor low battery",
			fmt.Sprintf("thermostat with low battery: %s/%s", tstat.ID, tstat.Name),
			[]string{"battery"})
	}

	// remove thermostats and humidistats whose last update is
	// greater than 6hr old. Stale data can cause the controller
	// to perform the wrong action.
	for id, tstat := range thermostats {
		if time.Since(tstat.Time) > 6*time.Hour {
			delete(thermostats, id)
		}
	}
	for id, hstat := range humidistats {
		if time.Since(hstat.Time) > 6*time.Hour {
			delete(humidistats, id)
		}
	}

	// update thermostats mapping
	if tstat.DewpointOnly {
		humidistats[tstat.ID] = tstat
	} else {
		thermostats[tstat.ID] = tstat
	}

	// find max and mean values
	var meanTemp float32 = 0.0
	var maxDewpoint float32 = 0
	var meanHeatSetpointErr float32 = 0.0
	var meanCoolSetpointErr float32 = 0.0

	for _, tstat := range thermostats {
		maxDewpoint = max(maxDewpoint, tstat.Dewpoint)
		meanTemp += tstat.Temperature
		meanHeatSetpointErr += tstat.Temperature - tstat.HeatSetpoint
		meanCoolSetpointErr += tstat.Temperature - tstat.CoolSetpoint
	}
	if len(thermostats) > 0 {
		meanTemp /= float32(len(thermostats))
		meanHeatSetpointErr /= float32(len(thermostats))
		meanCoolSetpointErr /= float32(len(thermostats))
	}

	for _, hstat := range humidistats {
		maxDewpoint = max(maxDewpoint, hstat.Dewpoint)
	}

	// update inputs and trigger controller routine
	inputs.Indoor.Temperature = meanTemp
	inputs.Indoor.Dewpoint = maxDewpoint
	inputs.Indoor.HeatSetpointErr = meanHeatSetpointErr
	inputs.Indoor.CoolSetpointErr = meanCoolSetpointErr

	// we need at least one thermostat to provide a room
	// temperature before the controller is ready
	if len(thermostats) > 0 {
		inputs.Ready |= IndoorReady
	}
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
	inputs.Outdoor.Humidity = data.RelHumidity
	inputs.Outdoor.Dewpoint = calculate_dewpoint_simple(data.Temperature, data.RelHumidity)
	inputs.Ready |= CurrentReady

	tryRunController(inputs)
}

func onAQHIUpdate(payload []byte) {
	var data weathergcca.AqhiForecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onAQHIUpdate:", err)
		return
	}
	if len(data.AQHI) == 0 {
		fmt.Println("bad data from AQHI update")
		return
	}

	inputMutex.Lock()
	defer inputMutex.Unlock()

	inputs.Outdoor.AQHI = int32(data.AQHI[0])
	inputs.Ready |= AQHIReady
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
