package main

import "time"

type HistoryData struct {
	SensorID    string
	Temperature float32
	Humidity    float32
	DewPoint    float32
	SetpointErr float32
	Time        time.Time
}
