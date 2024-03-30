package protocol

import "time"

type OutdoorConditions struct {
	OutdoorAirTemp    float32
	OutdoorAir24hLow  float32
	OutdoorAir24hAvg  float32
	OutdoorAir24hHigh float32
	LastUpdate        time.Time
}

type Thermostat struct {
	ID           string
	Name         string
	State        SensorData
	HeatSetpoint float32
	CoolSetpoint float32
}

type SensorData struct {
	Battery     int32
	LinkQuality int32
	Temperature float32
	Humidity    float32
	DewPoint    float32
	Time        time.Time
}
