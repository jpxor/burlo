package controller

import "time"

type Thermostat struct {
	ID   string
	Name string
	Time time.Time

	// sensors near the radiators and at floor level are
	// used to get accurate dewpoints to prevent condensation,
	// but they would report incorrect room temperature
	DewpointOnly bool

	Temperature  float32
	Humidity     float32
	Dewpoint     float32
	HeatSetpoint float32
	CoolSetpoint float32

	Battery     int32
	LinkQuality int32
}
