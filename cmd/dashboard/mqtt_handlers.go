package main

import (
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"burlo/pkg/weathergcca"
	"encoding/json"
	"fmt"
)

var dashboard = DashboardData{
	Thermostats: make(map[string]controller.Thermostat),
}

func onThermostatUpdate(payload []byte) {
	var tstat controller.Thermostat
	err := json.Unmarshal(payload, &tstat)
	if err != nil {
		fmt.Println("onThermostatUpdate:", err)
		return
	}
	dashboard.updateThermostat(tstat)
}

func onForecastUpdate(payload []byte) {
	var data weather.Forecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onForecastUpdate:", err)
		return
	}
	dashboard.updateTemperatureForcast(data)
}

func onCurrentWeatherUpdate(payload []byte) {
	var data weather.Current
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onCurrentWeatherUpdate:", err)
		return
	}
	dashboard.updateCurrentWeather(data)
}

func onAQHIUpdate(payload []byte) {
	var data weathergcca.AqhiForecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("onAQHIUpdate:", err)
		return
	}
	dashboard.updateAQHI(data)
}
