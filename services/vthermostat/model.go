package main

import (
	protocol "burlo/services/protocols"
	"time"
)

type HistoryData struct {
	SensorID    string
	Temperature float32
	Humidity    float32
	DewPoint    float32
	SetpointErr float32
	Time        time.Time
}

type Thermostat struct {
	ID           string
	Name         string
	Temperature  float32
	Humidity     float32
	DewPoint     float32
	HeatSetpoint float32
	CoolSetpoint float32
	Sensor       Sensor
	Time         time.Time
}

type Sensor struct {
	Battery     int32
	LinkQuality int32
}

func (t *Thermostat) From(s protocol.SensorData) {
	t.Temperature = s.Temperature
	t.Humidity = s.Humidity
	t.DewPoint = calculate_dewpoint_simple(t.Temperature, t.Humidity)
	t.Sensor.Battery = s.Battery
	t.Sensor.LinkQuality = s.LinkQuality
	t.Time = time.Now()
}

func calculate_dewpoint_simple(temp, relH float32) float32 {
	// a simple approximation, should err on the side of
	// being too high, never too low. Temperature must be
	// in celcius. Accureate to within 1 degC when RelH > 50%
	if relH >= 50 && temp >= 25 {
		return temp - ((100 - relH) / 5)
	} else {
		return temp - ((100 - relH) / 4)
	}
}
