package controller

import "time"

type Thermostat struct {
	ID   string
	Name string
	Time time.Time

	Temperature  float32
	Humidity     float32
	Dewpoint     float32
	HeatSetpoint float32
	CoolSetpoint float32

	Battery     int32
	LinkQuality int32
}
