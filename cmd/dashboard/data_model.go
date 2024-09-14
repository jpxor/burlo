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

func (t Temperature) asFloat(unit Unit) float32 {
	switch unit {
	case C:
		return float32(t)
	case F:
		return float32(t)*(9.0/5.0) + 32
	}
	panic(fmt.Sprintf("Temperature unit not implemented: '%s'", unit))
}

type DashboardData struct {
	Thermostats map[string]controller.Thermostat
	Weather     Weather
	Setpoint    SetpointData
	Unit        Unit
	Mutex       sync.Mutex
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

func (d *DashboardData) adjustSetpoint(adj float32) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	switch d.Setpoint.Mode {
	case Heat:
		d.Setpoint.HeatingSetpoint += Temperature(adj)
		pushSetpointToDashboards(d.Setpoint.HeatingSetpoint)
	case Cool:
		d.Setpoint.CoolingSetpoint += Temperature(adj)
		pushSetpointToDashboards(d.Setpoint.CoolingSetpoint)
	default:
		panic(fmt.Sprintf("mode not implemented: %s", d.Setpoint.Mode))
	}
}

func (d *DashboardData) updateThermostat(tstat controller.Thermostat) {
	d.Thermostats[tstat.ID] = tstat
	pushThermostatToDashboards(tstat)
}

func (d *DashboardData) updateTemperatureForcast(data weather.Forecast) {
	if len(data.Temperature) == 0 {
		fmt.Println("bad data from Temperature Forcast update")
		return
	}
	d.Weather.T24hHigh = Celcius(slices.Max(data.Temperature))
	d.Weather.T24hLow = Celcius(slices.Min(data.Temperature))
	d.Weather.T24hMean = Celcius(sliceMean(data.Temperature))
	d.Weather.Temperature = Celcius(data.Temperature[0])
	pushWeatherToDashboards(d.Weather)
}

func Celcius(v float32) Temperature {
	return Temperature(v)
}

func (d *DashboardData) updateCurrentWeather(data weather.Current) {
	d.Weather.Temperature = Celcius(data.Temperature)
	d.Weather.ConditionsCode = data.WeatherCode
	pushWeatherToDashboards(d.Weather)
}

func (d *DashboardData) updateAQHI(data weathergcca.AqhiForecast) {
	if len(data.AQHI) == 0 {
		fmt.Println("bad data from AQHI update")
		return
	}
	d.Weather.AirQualityIdx = int32(data.AQHI[0])
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
