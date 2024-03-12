package main

import (
	"fmt"
	"time"
)

type Mode int

const (
	Off Mode = iota
	On
	Heat
	Cool
	Auto
)

type SystemMode struct {
	Mode
	LastUpdate time.Time
}

type CirculatorState struct {
	Running    bool
	LastUpdate time.Time
}

type ControlState struct {
	CirculatorState
	SystemMode
	TargetSupplyTemperature float32
}

type IndoorConditions struct {
	SetpointError    float32
	DewPoint         float32
	IndoorAirTempMax float32
}

type OutdoorConditions struct {
	OutdoorAirTemp   float32
	OutdoorAir24hLow float32
	OutdoorAir24hAvg float32
}

type ControlConditions struct {
	IndoorConditions
	OutdoorConditions
}

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

// UpdateControls makes decisions based on outdoor temperature and humidity.
func UpdateControls(state ControlState, conditions ControlConditions) ControlState {

	state.SystemMode = UpdateSystemMode(state, conditions)
	state.TargetSupplyTemperature = UpdateSupplyWaterTemperature(state, conditions)
	state.CirculatorState = UpdateCirculatorState(state, conditions)

	fmt.Printf("System Mode: %+v\n", state.SystemMode)
	fmt.Printf("Supply Water Temperature: %.2f°C\n", state.TargetSupplyTemperature)
	fmt.Printf("Circulator Status: %v\n", state.CirculatorState)

	return state
}

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

// UpdateSystemMode handles auto-changeover for heat-off-cool modes
func UpdateSystemMode(state ControlState, conditions ControlConditions) SystemMode {

	// the web interface allows mannually setting 'heat', 'cool', 'off'
	if state.Mode == manual_override {
		return state.SystemMode
	}

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
		// decide when to turn off - the end of cooling season,
		if conditions.OutdoorAir24hAvg < cooling_mode_cutoff {
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

// UpdateSupplyWaterTemperature calculates the ideal supply water temperature (Celsius)
func UpdateSupplyWaterTemperature(state ControlState, conditions ControlConditions) float32 {
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
		return min(max_supply_temperature, max(min_heating_supply_temperature,
			target_supply_temperature))

	case Cool:
		// cooling temperature is set to ensure the floors don't
		// get uncomfortably cool. We can go lower during night,
		// But it must remain above dewpoint at all times!
		target_supply_temperature := comfort_cooling_supply_temperature
		if NightCoolingBoost() {
			target_supply_temperature = min_cooling_supply_temperature
		}
		return max(conditions.DewPoint+1.5, target_supply_temperature)

	default:
		// No supply water needed in Off mode, but return
		// a sane default anyway
		return design_indoor_air_temperature
	}
}

// UpdateCirculatorState determines whether the circulator should run.
func UpdateCirculatorState(state ControlState, conditions ControlConditions) CirculatorState {

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
			state.TargetSupplyTemperature < conditions.IndoorAirTempMax {
			return CirculatorState{
				Running:    false,
				LastUpdate: time.Now(),
			}
		}

	case Cool:
		if RoomTooCold(conditions.SetpointError) ||
			state.TargetSupplyTemperature > conditions.IndoorAirTempMax {
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
