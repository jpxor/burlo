package main

import "time"

type SensorData struct {
	Battery     int32
	LinkQuality int32
	Temperature float32
	Humidity    float32
	DewPoint    float32
	Time        time.Time
}

type Thermostat struct {
	ID          string
	Name        string
	State       SensorData
	Setpoint    float32
	SetpointErr float32
}

type HistoryData struct {
	SensorID    string
	Temperature float32
	Humidity    float32
	DewPoint    float32
	SetpointErr float32
	Time        time.Time
}
