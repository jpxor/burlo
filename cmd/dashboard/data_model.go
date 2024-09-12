package main

import (
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"burlo/pkg/weathergcca"
	"fmt"
	"slices"
)

type DashboardData struct {
	Thermostats map[string]controller.Thermostat
	Weather     Weather
}

type Weather struct {
	T24hHigh       float32
	T24hLow        float32
	T24hMean       float32
	Temperature    float32
	ConditionsCode int32
	AirQualityIdx  int32
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
	d.Weather.T24hHigh = slices.Max(data.Temperature)
	d.Weather.T24hLow = slices.Min(data.Temperature)
	d.Weather.T24hMean = sliceMean(data.Temperature)
	d.Weather.Temperature = data.Temperature[0]
	pushWeatherToDashboards(d.Weather)
}

func (d *DashboardData) updateCurrentWeather(data weather.Current) {
	d.Weather.Temperature = data.Temperature
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
