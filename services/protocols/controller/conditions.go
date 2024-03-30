package controller

import "time"

type Conditions struct {
	IndoorConditions
	OutdoorConditions
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
