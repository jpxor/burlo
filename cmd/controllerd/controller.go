package main

import (
	protocol "burlo/services/protocols"
	"fmt"
)

var currentState = CtrlOutput{
	DX2W:     DX2W{Mode: DX2W_AUTO},
	Window:   CLOSE,
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

	mode, state := selectDX2WMode(inputs, output)

	if inputs.ModeOverride != DX2W_AUTO {
		output.DX2W.Mode = inputs.ModeOverride
	} else {
		output.DX2W.setMode(mode)
	}

	if inputs.StateOverride != DX2W_STATE_AUTO {
		output.DX2W.State = inputs.StateOverride
	} else {
		output.DX2W.State = state
	}

	// order is important here, we need the dewpoint decide on
	// ventilation, and ventilation to
	output.Dewpoint = inputs.Indoor.Dewpoint

	window := selectWindowMode(inputs, output)
	if window != output.Window {
		output.Window = window
		if window == OPEN {
			notify.Publish(
				"Its nice out there!",
				"Now is a good time to open those windows and get some fresh air",
				[]string{"house_with_garden", "sun_behind_small_cloud"},
			)
		} else {
			notify.Publish(
				"Keep windows closed",
				fmt.Sprintf("Outdoors: %.1f°C and %.0f%% relH",
					inputs.Outdoor.Temperature,
					inputs.Outdoor.Humidity),
				[]string{"house_with_garden", "window"},
			)
		}
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
	// turn on heating mode if the house is losing heat and rooms are not too hot
	if inputs.Outdoor.T24hMean < 16 && !RoomTooHot(inputs.Indoor.HeatSetpointErr) {
		return DX2W_HEAT, DX2W_ON
	}
	// comfortable summer temp is 22degC
	// turn on cooling mode if house is gaining heat and rooms already too hot
	if inputs.Outdoor.T24hMean > 20 && !RoomTooCold(inputs.Indoor.CoolSetpointErr) {
		return DX2W_COOL, DX2W_ON
	}
	// OFF means the heatpump will stop maintaining the buffer temperature
	// which will save energy during long periods when heating/cooling is not
	// needed, especially when switching between heat/cool mode
	return current.DX2W.Mode, DX2W_OFF
}

func selectWindowMode(inputs CtrlInput, current CtrlOutput) wmode {
	// todo: keep windows closed if air quality is low

	switch current.DX2W.Mode {
	case DX2W_HEAT:
		// no need to worry about dewpoint in heating mode
		if inputs.Outdoor.Temperature >= 20 && inputs.Outdoor.Temperature <= 24 {
			return OPEN
		}
	case DX2W_COOL:
		// important to take dewpoint into account in cooling mode,
		// it can get very humid out during the summer
		if inputs.Outdoor.Dewpoint <= 12 || inputs.Outdoor.Dewpoint <= inputs.Indoor.Dewpoint {
			if inputs.Outdoor.Temperature <= 22 || inputs.Outdoor.Temperature < inputs.Indoor.Temperature-2 {
				if inputs.Indoor.Temperature >= 22 && inputs.Outdoor.Temperature >= 15 {
					return OPEN
				}
				if inputs.Outdoor.Temperature >= 18 {
					return OPEN
				}
			}
		}
	}
	return CLOSE
}

func updateZoneCalls(inputs CtrlInput, current CtrlOutput) bool {
	// no calls for heat/cool when the windows should be open instead
	// or if the system is off
	if current.Window == OPEN || current.DX2W.State == DX2W_OFF {
		return false
	}
	switch current.DX2W.Mode {
	case DX2W_HEAT:
		if !RoomTooHot(inputs.Indoor.HeatSetpointErr) {
			return true
		}
	case DX2W_COOL:
		if !RoomTooCold(inputs.Indoor.CoolSetpointErr) {
			return true
		}
	}
	// TODO: return true when the compressor is running

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
