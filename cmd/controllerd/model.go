package main

type wmode string
type bitflag uint8

const IndoorReady bitflag = 0b001
const CurrentReady bitflag = 0b010
const ForecastReady bitflag = 0b100
const AQHIReady bitflag = 0b1000

type State struct {
	Inputs  CtrlInput
	Outputs CtrlOutput
}

type CtrlInput struct {
	Ready  bitflag
	Indoor struct {
		Temperature     float32
		Dewpoint        float32
		HeatSetpointErr float32
		CoolSetpointErr float32
	}
	Outdoor struct {
		Temperature float32
		Humidity    float32
		Dewpoint    float32
		T24hHigh    float32
		T24hLow     float32
		T24hMean    float32
		AQHI        int32
	}
	ModeOverride  dx2wmode
	StateOverride dx2wstate
}

var OPEN wmode = "OPEN"
var CLOSE wmode = "CLOSE"

type CtrlOutput struct {
	DX2W     DX2W
	Window   wmode
	Dewpoint float32
	ZoneCall bool
}

func modelHeatLoad(outdoorTemp, indoorTemp float32) float32 {
	const factor = 1000
	dT := outdoorTemp - indoorTemp
	return dT * factor
}
