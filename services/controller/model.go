package main

import "time"

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
}

type OutdoorConditions struct {
	OutdoorAirTemp   float32
	OutdoorAir24hLow float32
	OutdoorAir24hAvg float32
}

type ControlConditions struct {
	IndoorConditions
	OutdoorConditions
}

type system_state struct {
	ControlState
	ControlConditions
	Thermostats map[string]Thermostat
}
