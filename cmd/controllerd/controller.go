package main

import (
	protocol "burlo/services/protocols"
)

var currentState = CtrlOutput{
	DX2W:     DX2W{Mode: DX2W_AUTO},
	Window:   CLOSE,
	Dewpoint: 0,
	ZoneCall: false,
}

func tryRunController(inputs CtrlInput) {
	if inputs.Ready != (IndoorReady | CurrentReady | ForecastReady | AQHIReady) {
		return
	}
	runController(inputs)
}

func runController(inputs CtrlInput) {
	var output CtrlOutput = currentState

	mode, state := selectDX2WMode(inputs, output)

	if inputs.ModeOverride != DX2W_AUTO {
		output.DX2W.Mode = inputs.ModeOverride
	} else {
		output.DX2W.setMode(mode)
	}

	if inputs.StateOverride != DX2W_STATE_AUTO {
		output.DX2W.State = inputs.StateOverride
	} else {
		output.DX2W.setState(state)
	}

	// order is important here, we need the dewpoint decide on
	// ventilation, and ventilation to
	output.Dewpoint = inputs.Indoor.Dewpoint

	window := selectWindowMode(inputs, output)
	if window != output.Window {
		output.Window = window
		notifyWindow(window)
	}

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
	// TODO: use modbus to set dx2w state (on/off)
	currentState = output
}

func selectDX2WMode(inputs CtrlInput, current CtrlOutput) (dx2wmode, dx2wstate) {
	// zero load outdoor temp is 16degC
	coldOut := inputs.Outdoor.T24hMean < 16 && inputs.Outdoor.T24hHigh < 20
	hotOut := inputs.Outdoor.T24hMean > 20 && inputs.Outdoor.T24hLow > 16

	coolSetpoint := inputs.Indoor.Temperature - inputs.Indoor.CoolSetpointErr
	heatSetpoint := inputs.Indoor.Temperature - inputs.Indoor.HeatSetpointErr
	midpoint := (coolSetpoint + heatSetpoint) / 2

	// set initial mode when auto
	// this only runs once after the controller is started
	currentMode := current.DX2W.Mode
	if currentMode == DX2W_AUTO {
		if coldOut {
			currentMode = DX2W_HEAT
		} else {
			currentMode = DX2W_COOL
		}
	}

	switch {
	case coldOut:
		// maintain heat mode if its cold out
		// switch to heat mode if its cold out and rooms are near the heat setpoint
		if currentMode == DX2W_HEAT || inputs.Indoor.Temperature < midpoint {
			return DX2W_HEAT, DX2W_ON
		}
		// turn off when in cool mode and its cold out
		return currentMode, DX2W_OFF

	case hotOut:
		// maintain cool mode if its hot out
		// switch to cool mode if its hot out and rooms are near the cool setpoint
		if currentMode == DX2W_COOL || inputs.Indoor.Temperature > midpoint {
			return DX2W_COOL, DX2W_ON
		}
		// turn off when in heat mode and its hot out
		return currentMode, DX2W_OFF

	default: // maintain current mode during mild weather
		return currentMode, DX2W_ON
	}
}

func selectWindowMode(inputs CtrlInput, current CtrlOutput) wmode {
	// keep windows closed if air quality health risk is Moderate to high
	// Risk: Low (1-3)	Moderate (4-6)	High (7-10)	Very high (above 10)
	// TODO: make configurable
	if inputs.Outdoor.AQHI > 5 {
		return CLOSE
	}
	switch current.DX2W.Mode {
	case DX2W_HEAT:
		if inputs.Outdoor.Temperature < (16+20)/2 {
			return CLOSE
		}
		// edge case: it can get hot and humid out before the system switches
		// over to COOL mode, should keep windows closed then too
		if inputs.Outdoor.Temperature > 24 && inputs.Outdoor.Dewpoint > 16 {
			return CLOSE
		}
		return OPEN

	case DX2W_COOL:
		dT := inputs.Outdoor.Temperature - inputs.Indoor.Temperature
		coolSetpoint := inputs.Indoor.Temperature - inputs.Indoor.CoolSetpointErr
		heatSetpoint := inputs.Indoor.Temperature - inputs.Indoor.HeatSetpointErr

		midpoint := (coolSetpoint + heatSetpoint) / 2
		lowpoint := heatSetpoint + min(0, heatSetpoint-inputs.Indoor.Temperature) - 1

		outdoorTempLow := inputs.Outdoor.Temperature < lowpoint
		outdoorTempHigh := inputs.Outdoor.Temperature > coolSetpoint

		// important to take dewpoint into account in cooling mode,
		// it can get very humid out during the summer
		dewpointLow := inputs.Outdoor.Dewpoint <= 12 || inputs.Outdoor.Dewpoint <= inputs.Indoor.Dewpoint
		dewpointOk := inputs.Outdoor.Dewpoint < 16 && inputs.Outdoor.Dewpoint < inputs.Indoor.Dewpoint+1
		dewpointCheck := dewpointLow || (dewpointOk && dT < -4)

		// keep windows closed if dewpoint is too high
		if !dewpointCheck {
			return CLOSE
		}

		// keep windows closed if outdoor temp is too high
		if outdoorTempHigh && dT > -2 {
			return CLOSE
		}

		// keep windows closed if outdoor temp is too low
		if outdoorTempLow {
			return CLOSE
		}

		// open windows to help cool
		if dT < 0 && inputs.Indoor.Temperature >= midpoint {
			return OPEN
		}

		// open the windows because its nice out
		return OPEN

	default:
		return CLOSE
	}
}

func updateZoneCalls(inputs CtrlInput, current CtrlOutput) bool {
	if current.Window == OPEN || current.DX2W.State == DX2W_OFF {
		return false
	}
	switch current.DX2W.Mode {
	case DX2W_HEAT:
		condA := RoomTooCold(inputs.Indoor.HeatSetpointErr) && inputs.Outdoor.Temperature < 20
		condB := !RoomTooHot(inputs.Indoor.HeatSetpointErr) && inputs.Outdoor.Temperature < 16
		return condA || condB

	case DX2W_COOL:
		condA := RoomTooHot(inputs.Indoor.CoolSetpointErr) && inputs.Outdoor.Temperature > 16
		condB := !RoomTooCold(inputs.Indoor.CoolSetpointErr) && inputs.Outdoor.Temperature > 20
		return condA || condB

	default:
		return false
	}
}

// RoomTooCold is a helper that returns true if the
// room temperature falls below the target (with some margin)
func RoomTooCold(setpointErr float32) bool {
	return setpointErr < -0.5 // example: {target=20, too_cold=19.5}
}

// RoomTooHot is a helper that returns true if the
// room temperature rises above the target (with some margin)
func RoomTooHot(setpointErr float32) bool {
	return setpointErr > 0.5 // example: {target=20, too_hot=20.5}
}

func belowSetpoint(setpointErr float32) bool {
	return setpointErr < 0
}

func aboveSetpoint(setpointErr float32) bool {
	return setpointErr > 0
}
