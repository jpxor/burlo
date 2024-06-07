package main

import (
	"fmt"
	"time"
)

type dx2wmode string
type dx2wstate string

var DX2W_AUTO dx2wmode = "AUTO"
var DX2W_HEAT dx2wmode = "HEAT"
var DX2W_COOL dx2wmode = "COOL"

var DX2W_ON dx2wstate = "ON"
var DX2W_OFF dx2wstate = "OFF"
var DX2W_STATE_AUTO dx2wstate = "AUTO"

type DX2W struct {
	Mode       dx2wmode
	State      dx2wstate
	LastChange time.Time
}

func (dx2w *DX2W) setMode(mode dx2wmode) {
	if mode == dx2w.Mode {
		return
	}
	// debounce when setting heat or cool
	if (mode == DX2W_HEAT || mode == DX2W_COOL) &&
		time.Since(dx2w.LastChange) < 24*time.Hour {
		return
	}
	dx2w.Mode = mode
	dx2w.LastChange = time.Now()

	if mode == DX2W_HEAT {
		notify.Publish(
			"Heating mode activated",
			"Its getting chilly out there",
			[]string{"house_with_garden", "fire"},
		)
	} else {
		notify.Publish(
			"Cooling mode activated",
			"Wow its hot out there",
			[]string{"house_with_garden", "snowflake"},
		)
	}
}

func (dx2w DX2W) String() string {
	return fmt.Sprintf("DX2W_Mode_%s_%s", dx2w.Mode, dx2w.State)
}

func dewpointToVoltage(temperature float32) float32 {
	convertToFahrenheit := func(celsius float32) float32 {
		return celsius*9.0/5.0 + 32.0
	}
	temperature = convertToFahrenheit(temperature)
	var x1, y1 float32 = 94.4, 10.0
	var x2, y2 float32 = 32.0, 0.00
	m := (y2 - y1) / (x2 - x1)
	b := y1 - m*x1
	return m*temperature + b
}
