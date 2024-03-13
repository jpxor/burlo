package main

import (
	"fmt"
	"time"
)

var manual_override Mode = Auto

// configs
var min_cooling_supply_temperature float32 = 12
var comfort_cooling_supply_temperature float32 = 18
var max_supply_temperature float32 = 40.55
var design_supply_temperature float32 = 40.55
var design_outdoor_air_temperature float32 = -25
var design_indoor_air_temperature float32 = 20
var zero_load_outdoor_air_temperature float32 = 16
var cooling_mode_cutoff float32 = 20

// RoomTooCold is a helper that returns true if the
// room temperature falls below the target (with some margin)
func RoomTooCold(setpoint_error float32) bool {
	return setpoint_error <= -1 // example: {target=20, too_cold=19.0}
}

// RoomTooHot is a helper that returns true if the
// room temperature rises above the target (with some margin)
func RoomTooHot(setpoint_error float32) bool {
	return setpoint_error >= 1 // example: {target=20, too_hot=21}
}

// UpdateControls
func UpdateControls(state ControlState, conditions ControlConditions) ControlState {
	state.SystemMode = updateSystemMode(state, conditions)
	state.SupplyTemperature = updateSupplyWaterTemperature(state, conditions)
	state.CirculatorState = updateCirculatorState(state, conditions)
	return state
}

// updateSystemMode handles auto-changeover for heat-off-cool modes
func updateSystemMode(state ControlState, conditions ControlConditions) SystemMode {

	// debounce, don't let the mode switch too often
	if time.Since(state.SystemMode.LastUpdate) < 24*time.Hour {
		return state.SystemMode
	}

	switch state.Mode {
	case Heat:
		// decide when to turn off - the end of heating season,
		// can be conservative since there is nearly no cost to
		// maintaining heating mode while not heating (keeps buffer
		// near room temperature when its warm out)
		if conditions.OutdoorAir24hLow > zero_load_outdoor_air_temperature {
			return SystemMode{
				Off, time.Now(),
			}
		}

	case Cool:
		// decide when to turn off - don't want this running too
		// often, so we are aggresive when turning it off
		if conditions.OutdoorAir24hAvg < cooling_mode_cutoff &&
			!RoomTooHot(conditions.SetpointError) {
			return SystemMode{
				Off, time.Now(),
			}
		}

	case Off:
		// if the average outdoor temperature is higher than our cooling setpoint, then
		// the room temperatures will start to rise above that setpoint, which triggers
		// cooling mode
		if (conditions.OutdoorAir24hAvg > design_indoor_air_temperature) &&
			RoomTooHot(conditions.SetpointError) {
			return SystemMode{
				Cool, time.Now(),
			}
		}
		// if the average outdoor temp is less than the zero-load, then room temperatures will
		// start to drop until they become too cold, which triggers heating mode
		if (conditions.OutdoorAir24hAvg < zero_load_outdoor_air_temperature) &&
			RoomTooCold(conditions.SetpointError) {
			return SystemMode{
				Heat, time.Now(),
			}
		}
	}

	// no change
	return state.SystemMode
}

// updateSupplyWaterTemperature calculates the ideal supply water temperature (Celsius)
func updateSupplyWaterTemperature(state ControlState, conditions ControlConditions) SupplyTemperature {
	switch state.Mode {
	case Heat:
		// linear relationship from no-load to design-load
		// re: Heating Load Line Chart
		min_heating_supply_temperature := design_indoor_air_temperature
		m := (design_supply_temperature - min_heating_supply_temperature) /
			(design_outdoor_air_temperature - zero_load_outdoor_air_temperature)
		b := design_supply_temperature - (m * design_outdoor_air_temperature)
		t := conditions.OutdoorAirTemp
		target_supply_temperature := m*t + b

		// correction with delay
		if time.Since(state.SupplyTemperature.LastUpdate) > 15*time.Minute {
			if RoomTooHot(conditions.SetpointError) {
				state.SupplyTemperature.Correction -= 1
				state.SupplyTemperature.LastUpdate = time.Now()
			}
			if RoomTooCold(conditions.SetpointError) {
				state.SupplyTemperature.Correction += 1
				state.SupplyTemperature.LastUpdate = time.Now()
			}
		}
		target_supply_temperature += state.SupplyTemperature.Correction

		state.SupplyTemperature.Target = min(max_supply_temperature, max(min_heating_supply_temperature, target_supply_temperature))
		return state.SupplyTemperature

	case Cool:
		// cooling temperature is set to ensure the floors don't
		// get uncomfortably cool. We can go lower during night,
		// But it must remain above dewpoint at all times!
		target_supply_temperature := comfort_cooling_supply_temperature
		if NightCoolingBoost() {
			target_supply_temperature = min_cooling_supply_temperature
		}
		state.SupplyTemperature.Target = max(conditions.DewPoint+1.5, target_supply_temperature)
		return state.SupplyTemperature

	default:
		// No supply water needed in Off mode, but return
		// a sane default anyway
		state.SupplyTemperature.Target = design_indoor_air_temperature
		return state.SupplyTemperature
	}
}

// updateCirculatorState determines whether the circulator should run.
func updateCirculatorState(state ControlState, conditions ControlConditions) CirculatorState {

	// debounce, don't let the circulator switch too often
	if state.CirculatorState.Running {
		// run for at least 15 mins
		if time.Since(state.CirculatorState.LastUpdate) < 15*time.Minute {
			return state.CirculatorState
		}
	} else {
		// stay off for at least 1 min
		if time.Since(state.CirculatorState.LastUpdate) < 1*time.Minute {
			return state.CirculatorState
		}
	}

	switch state.Mode {
	case Heat:
		if RoomTooHot(conditions.SetpointError) ||
			state.SupplyTemperature.Target < conditions.IndoorAirTempMax {
			return CirculatorState{
				Running:    false,
				LastUpdate: time.Now(),
			}
		}

	case Cool:
		if RoomTooCold(conditions.SetpointError) ||
			state.SupplyTemperature.Target > conditions.IndoorAirTempMax {
			return CirculatorState{
				Running:    false,
				LastUpdate: time.Now(),
			}
		}

	case Off:
		return CirculatorState{
			Running:    false,
			LastUpdate: time.Now(),
		}
	}

	// no change if already running
	if state.CirculatorState.Running {
		return state.CirculatorState
	}

	return CirculatorState{
		Running:    true,
		LastUpdate: time.Now(),
	}
}

func NightCoolingBoost() bool {
	///////////
	layout := "15:04"
	startTime, err := time.Parse(layout, "23:59")
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		return false
	}
	endTime, err := time.Parse(layout, "4:00")
	if err != nil {
		fmt.Println("Error parsing end time:", err)
		return false
	}
	/////////
	now := time.Now()
	if endTime.Before(startTime) {
		return now.After(startTime) || now.Before(endTime)
	}
	return now.After(startTime) && now.Before(endTime)
}

func sample() {
	// Example usage:
	conditions := ControlConditions{
		IndoorConditions{
			SetpointError:    0,
			DewPoint:         12,
			IndoorAirTempMax: 21,
		},
		OutdoorConditions{
			OutdoorAirTemp:   -15,
			OutdoorAir24hAvg: -10,
			OutdoorAir24hLow: -20,
		},
	}
	mode := SystemMode{
		Mode: Heat,
	}

	constrolState := ControlState{
		SystemMode: mode,
	}
	UpdateControls(constrolState, conditions)
}
