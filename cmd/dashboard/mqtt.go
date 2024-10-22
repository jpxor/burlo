package main

import (
	"burlo/config"
	"burlo/pkg/models/controller"
	"burlo/pkg/models/weather"
	"burlo/pkg/mqtt"
	"burlo/pkg/weathergcca"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (d *Dashboard) mqttListener(ctx context.Context, cfg config.ServiceConf) {
	mqtt.NewClient(mqtt.Opts{
		Context:  ctx,
		Address:  cfg.Mqtt.Address,
		User:     cfg.Mqtt.User,
		Pass:     []byte(cfg.Mqtt.Pass),
		ClientID: "dashboard-listener",
		Topics: []string{
			"burlo/controller/thermostats/#",
			"burlo/controller/humidistat/#",
			"burlo/controller/setpoints/selected_tstat",
			"burlo/controller/setpoints/heating",
			"burlo/controller/setpoints/cooling",
			"burlo/controller/setpoints/mode",
			"burlo/weather/current",
			"burlo/weather/forecast",
			"burlo/weather/aqhi",
		},
		OnPublishRecv: d.handleMqttUpdates,
	})
	// waits for signal
	<-ctx.Done()
}

func (d *Dashboard) handleMqttUpdates(topic string, payload []byte) {
	topic = strings.TrimPrefix(topic, "burlo/")

	switch {
	case strings.HasPrefix(topic, "controller/thermostats/"):
		d.onMqttThermostatsUpdate(payload)

	case strings.HasPrefix(topic, "weather/current"):
		d.onMqttCurrentWeatherUpdate(payload)

	case strings.HasPrefix(topic, "weather/forecast"):
		d.onMqttForecastUpdate(payload)

	case strings.HasPrefix(topic, "weather/aqhi"):
		d.onMqttAQHIUpdate(payload)

	case strings.HasPrefix(topic, "controller/setpoints/"):
		topic = strings.TrimPrefix(topic, "controller/setpoints/")
		d.onMqttSetpointsUpdate(topic, payload)

	default:
		fmt.Println("unhandled topic:", topic)
	}
}

func (d *Dashboard) onMqttThermostatsUpdate(payload []byte) {
	var tstat controller.Thermostat
	err := json.Unmarshal(payload, &tstat)
	if err != nil {
		fmt.Println("ERROR onMqttThermostatsUpdate: invalid thermostat data:", err, string(payload))
		return
	}
	d.setThermostat(tstat)
}

func (d *Dashboard) onMqttForecastUpdate(payload []byte) {
	var data weather.Forecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("ERROR onForecastUpdate:", err)
		return
	}
	d.updateTemperatureForcast(data)
}

func (d *Dashboard) onMqttCurrentWeatherUpdate(payload []byte) {
	var data weather.Current
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("ERROR onCurrentWeatherUpdate:", err)
		return
	}
	d.updateCurrentWeather(data)
}

func (d *Dashboard) onMqttAQHIUpdate(payload []byte) {
	var data weathergcca.AqhiForecast
	err := json.Unmarshal(payload, &data)
	if err != nil {
		fmt.Println("ERROR onAQHIUpdate:", err)
		return
	}
	d.updateAQHI(data)
}

func (d *Dashboard) onMqttSetpointsUpdate(topic string, payload []byte) {
	switch topic {
	case "selected_tstat":
		tstatName := string(payload)
		d.setPrimaryThermostat(tstatName)

	case "heating":
		val_celcius, err := strconv.ParseFloat(string(payload), 32)
		if err != nil {
			fmt.Println("ERROR handleMqttBurloData: heating setpoint not a number:", string(payload))
			return
		}
		d.setHeatingSetpoint(float32(val_celcius))

	case "cooling":
		val_celcius, err := strconv.ParseFloat(string(payload), 32)
		if err != nil {
			fmt.Println("ERROR handleMqttBurloData: cooling setpoint not a number:", string(payload))
			return
		}
		d.setCoolingSetpoint(float32(val_celcius))

	case "mode":
		mode := string(payload)
		d.setSetpointMode(mode)

	}
}

///////////////////////////////////////////////////////////
//
//   Server calls the following funcs to publish to mqtt
//
///////////////////////////////////////////////////////////

// adjusts the current mode's setpoint
func (s *DashboardServer) adjustSetpoint(adj float32) {
	validSetpointTemperature := func(t Temperature) bool {
		tc := t.asFloat(C)
		return tc >= 10 && tc <= 35
	}

	s.dashboard.Mutex.RLock()
	defer s.dashboard.Mutex.RUnlock()

	switch s.dashboard.Setpoint.Mode {
	case Heat:
		setpoint := s.dashboard.Setpoint.HeatingSetpoint + Temperature(adj)
		if !validSetpointTemperature(setpoint) {
			return
		}
		s.mqttc.Publish(true, "burlo/controller/setpoints/heating", setpoint.asFloat(C))

	case Cool:
		setpoint := s.dashboard.Setpoint.CoolingSetpoint + Temperature(adj)
		if !validSetpointTemperature(setpoint) {
			return
		}
		s.mqttc.Publish(true, "burlo/controller/setpoints/cooling", setpoint.asFloat(C))

	default:
		panic(fmt.Errorf("mode not implemented: %v", s.dashboard.Setpoint.Mode))
	}
}
