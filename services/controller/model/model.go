package model

import (
	services "burlo/services/model"
	"time"
)

type Mode int

const (
	Off Mode = iota
	On
	Heat
	Cool
	Auto
)

type SystemMode struct {
	Mode
	LastUpdate time.Time
}

type CirculatorState struct {
	Running    bool
	LastUpdate time.Time
}

type SupplyTemperature struct {
	Target     float32
	Correction float32
	LastUpdate time.Time
}

type ControlState struct {
	CirculatorState
	SystemMode
	SupplyTemperature
}

type IndoorConditions struct {
	SetpointError    float32
	DewPoint         float32
	IndoorAirTempMax float32
	LastUpdate       time.Time
}

type OutdoorConditions struct {
	OutdoorAirTemp    float32
	OutdoorAir24hLow  float32
	OutdoorAir24hAvg  float32
	OutdoorAir24hHigh float32
	LastUpdate        time.Time
}

type ControlConditions struct {
	IndoorConditions
	OutdoorConditions
}

type SystemState struct {
	ControlState
	ControlConditions
	Thermostats map[string]services.Thermostat
}
