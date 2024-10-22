package main

import (
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"burlo/pkg/weathergcca"
	"fmt"
	"slices"
	"sync"
)

type Mode string

var Heat Mode = "heat"
var Cool Mode = "cool"

type Unit string

var C Unit = "C"
var F Unit = "F"

type Temperature float32

func Celcius(v float32) Temperature {
	return Temperature(v)
}

func (t Temperature) asFloat(unit Unit) float32 {
	switch unit {
	case C:
		return float32(t)
	case F:
		return float32(t)*(9.0/5.0) + 32
	}
	panic(fmt.Sprintf("Temperature unit not implemented: '%s'", unit))
}

type Dashboard struct {
	Thermostats map[string]controller.Thermostat
	Weather     Weather
	Setpoint    SetpointData
	Unit        Unit
	Mutex       sync.RWMutex
}

type Weather struct {
	T24hHigh       Temperature
	T24hLow        Temperature
	T24hMean       Temperature
	Temperature    Temperature
	ConditionsCode int32
	AirQualityIdx  int32
}

type SetpointData struct {
	PrimaryThermostat string // name
	HeatingSetpoint   Temperature
	CoolingSetpoint   Temperature
	Mode              Mode
}

func NewDashboard() Dashboard {
	return Dashboard{
		Thermostats: make(map[string]controller.Thermostat),
		Setpoint: SetpointData{
			Mode:            Heat,
			HeatingSetpoint: 20,
			CoolingSetpoint: 24,
		},
		Unit: C,
	}
}

///////////////////////////////////////////////////////////
//
//   mqtt listener client calls the following funcs to update
//   the dashboard when changes are published
//
///////////////////////////////////////////////////////////

func (d *Dashboard) setPrimaryThermostat(name string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if _, ok := d.Thermostats[name]; ok {
		d.Setpoint.PrimaryThermostat = name
	} else {
		fmt.Println("WARN setPrimaryThermostat: unknown name:", name)
	}
}

func (d *Dashboard) setHeatingSetpoint(val_celcius float32) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if val_celcius < 10 && val_celcius > 40 {
		fmt.Println("WARN setHeatingSetpoint: invalid setpoint:", val_celcius, "degC")
		return
	}
	d.Setpoint.HeatingSetpoint = Celcius(val_celcius)
	pushSetpointToDashboards(d.Setpoint)
}

func (d *Dashboard) setCoolingSetpoint(val_celcius float32) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if val_celcius < 10 && val_celcius > 40 {
		fmt.Println("WARN setCoolingSetpoint: invalid setpoint:", val_celcius, "degC")
		return
	}
	d.Setpoint.CoolingSetpoint = Celcius(val_celcius)
	pushSetpointToDashboards(d.Setpoint)
}

func (d *Dashboard) setSetpointMode(modestr string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	mode := Mode(modestr)
	if mode != Heat && mode != Cool {
		fmt.Println("WARN setSetpointMode: unknown mode:", modestr)
		return
	}
	d.Setpoint.Mode = mode
	pushSetpointToDashboards(d.Setpoint)
}

func (d *Dashboard) setThermostat(tstat controller.Thermostat) {
	d.Mutex.Lock()
	d.Thermostats[tstat.ID] = tstat
	d.Mutex.Unlock()
	pushThermostatToDashboards(tstat)
}

func (d *Dashboard) updateTemperatureForcast(data weather.Forecast) {
	if len(data.Temperature) == 0 {
		fmt.Println("ERROR bad data from Temperature Forcast update")
		return
	}
	d.Mutex.Lock()
	d.Weather.T24hHigh = Celcius(slices.Max(data.Temperature))
	d.Weather.T24hLow = Celcius(slices.Min(data.Temperature))
	d.Weather.T24hMean = Celcius(sliceMean(data.Temperature))
	d.Weather.Temperature = Celcius(data.Temperature[0])
	d.Mutex.Unlock()
	pushWeatherToDashboards(d.Weather)
}

func (d *Dashboard) updateCurrentWeather(data weather.Current) {
	d.Mutex.Lock()
	d.Weather.Temperature = Celcius(data.Temperature)
	d.Weather.ConditionsCode = data.WeatherCode
	d.Mutex.Unlock()
	pushWeatherToDashboards(d.Weather)
}

func (d *Dashboard) updateAQHI(data weathergcca.AqhiForecast) {
	if len(data.AQHI) == 0 {
		fmt.Println("ERROR bad data from AQHI update")
		return
	}
	d.Mutex.Lock()
	d.Weather.AirQualityIdx = int32(data.AQHI[0])
	d.Mutex.Unlock()
	pushWeatherToDashboards(d.Weather)
}

func sliceMean(list []float32) float32 {
	var sum float32
	for _, f := range list {
		sum += f
	}
	return sum / float32(len(list))
}

// func calculate_dewpoint_simple(temp, relH float32) float32 {
// 	if relH >= 50 && temp >= 25 {
// 		return temp - ((100 - relH) / 5)
// 	} else {
// 		return temp - ((100 - relH) / 4)
// 	}
// }
