package main

import (
	protocol "burlo/services/protocols"
)

var currentState = CtrlOutput{
	DX2W:     DX2W{Mode: DX2W_AUTO},
	Window:   WCLOSE,
	Dewpoint: 0,
	ZoneCall: false,
}

func tryRunController(inputs CtrlInput) {
	if inputs.Ready != (IndoorReady | CurrentReady | ForecastReady) {
		return
	}
	runController(inputs)
}

func runController(inputs CtrlInput) {
	var output CtrlOutput = currentState

	if inputs.ModeOverride != DX2W_AUTO {
		output.DX2W.Mode = inputs.ModeOverride
	} else {
		selected := selectDX2WMode(inputs, output)
		output.DX2W.set(selected)
	}
	// order is important here, we need the dewpoint decide on
	// ventilation, and ventilation to
	output.Dewpoint = inputs.Indoor.Dewpoint
	output.Window = selectWindowMode(inputs, output)
	output.ZoneCall = updateZoneCalls(inputs, output)

	// apply new state
	set_digital_out(protocol.PhidgetDO{
		Name:    "CoolingMode",
		HubPort: 0,
		Channel: 1,
		Output:  output.DX2W.Mode == DX2W_COOL,
	})
	set_digital_out(protocol.PhidgetDO{
		Name:    "ZoneCirculator",
		HubPort: 0,
		Channel: 0,
		Output:  output.ZoneCall,
	})
	set_voltage_out(protocol.PhidgetVO{
		Name:    "Dewpoint",
		HubPort: 1,
		Channel: 0,
		Output:  dewpointToVoltage(output.Dewpoint),
	})
	currentState = output
}

func selectDX2WMode(inputs CtrlInput, _ CtrlOutput) dx2wmode {
	// zero load outdoor temp is 16degC
	// turn on heating mode if the house is losing heat and rooms are not too hot
	if inputs.Outdoor.T24hMean < 16 && !RoomTooHot(inputs.Indoor.HeatSetpointErr) {
		return DX2W_HEAT
	}
	// comfortable summer temp is 22degC
	// turn on cooling mode if house is gaining heat and rooms already too hot
	if inputs.Outdoor.T24hMean > 20 && !RoomTooCold(inputs.Indoor.CoolSetpointErr) {
		return DX2W_COOL
	}
	return DX2W_OFF
}

func selectWindowMode(inputs CtrlInput, current CtrlOutput) wmode {
	// todo: keep windows closed if air quality is low

	switch current.DX2W.Mode {
	case DX2W_HEAT:
		// no need to worry about dewpoint in heating mode
		if inputs.Outdoor.Temperature >= 20 && inputs.Outdoor.Temperature <= 24 {
			return WOPEN
		}
	case DX2W_COOL:
	case DX2W_OFF:
		// important to take dewpoint into account in cooling mode,
		// it can get very humid out during the summer
		if inputs.Outdoor.Dewpoint <= 12 || inputs.Outdoor.Dewpoint <= inputs.Indoor.Dewpoint {
			if inputs.Outdoor.Temperature <= 22 || inputs.Outdoor.Temperature < inputs.Indoor.Temperature-2 {
				if inputs.Indoor.Temperature >= 22 && inputs.Outdoor.Temperature >= 15 {
					return WOPEN
				}
				if inputs.Outdoor.Temperature >= 18 {
					return WOPEN
				}
			}
		}
	}
	return WCLOSE
}

func updateZoneCalls(inputs CtrlInput, current CtrlOutput) bool {
	// no calls for heat/cool when the windows should be open instead
	if current.Window == WOPEN {
		return false
	}
	switch current.DX2W.Mode {
	case DX2W_OFF:
		return false

	case DX2W_HEAT:
		if !RoomTooHot(inputs.Indoor.HeatSetpointErr) {
			return true
		}

	case DX2W_COOL:
		if !RoomTooCold(inputs.Indoor.CoolSetpointErr) {
			return true
		}
	}
	// default to off
	return false
}

// RoomTooCold is a helper that returns true if the
// room temperature falls below the target (with some margin)
func RoomTooCold(setpoint_error float32) bool {
	return setpoint_error < -0.5 // example: {target=20, too_cold=19.5}
}

// RoomTooHot is a helper that returns true if the
// room temperature rises above the target (with some margin)
func RoomTooHot(setpoint_error float32) bool {
	return setpoint_error > 0.5 // example: {target=20, too_hot=20.5}
}
