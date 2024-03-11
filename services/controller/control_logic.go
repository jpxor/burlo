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

// configs
var min_cooling_supply_temperature float32 = 12
var comfort_cooling_supply_temperature float32 = 18
var max_supply_temperature float32 = 40.55
var design_supply_temperature float32 = 40.55
var design_outdoor_air_temperature float32 = -25
var target_indoor_air_temperature float32 = 20
var zero_load_outdoor_air_temperature float32 = 16
var manual_override Mode = Auto
var cooling_setpoint float32 = 24

// measurements
var outdoor_air_temp float32           // updated every 15 mins
var outdoor_air_temp_trailing float32  // updated every 15 mins, average over the previous 24hrs
var max_indoor_air_temperature float32 // updated in realtime as temperatues change
var avg_setpoint_err float32           // updated in realtime as temperatues change
var dewPoint float32                   // updated in realtime as temperatures and humidity change
var min_outdoor_air_temp float32       // past 24hr & forcasted next 2 days

// DecisionLogic makes decisions based on outdoor temperature and humidity.
func DecisionLogic() {

	mode := SystemMode{
		Mode: Off,
	}

	mode = UpdateSystemMode(mode,
		outdoor_air_temp_trailing,
		outdoor_air_temp,
		avg_setpoint_err)

	supplyTemp := UpdateSupplyWaterTemperature(mode, outdoor_air_temp, dewPoint)
	circulatorState := UpdateCirculatorState(mode, supplyTemp, avg_setpoint_err)

	fmt.Printf("System Mode: %+v\n", mode)
	fmt.Printf("Supply Water Temperature: %.2f°C\n", supplyTemp)
	fmt.Printf("Circulator Status: %v\n", circulatorState)
}

// RoomTooCold is a helper that returns true if the
// room temperature falls below the target (with some margin)
func RoomTooCold(setpoint_error float32) bool {
	return setpoint_error <= -1 // example: {target=20, too_cold=19.0}
}

// RoomTooHot is a helper that returns true if the
// room temperature rises above the target (with some margin)
func RoomTooHot(setpoint_error float32) bool {
	return setpoint_error > 2 // example: {target=20, too_hot=22}
}

// UpdateSystemMode handles auto-changeover for heat-off-cool modes
func UpdateSystemMode(mode SystemMode, avg_outdoor_air_temp, outdoor_air_temp, avg_setpoint_err float32) SystemMode {
	if manual_override != Auto {
		return SystemMode{
			Mode: manual_override,
		}
	}

	if time.Since(mode.LastUpdate) < 24*time.Hour {
		return mode
	}

	switch mode.Mode {
	case Heat:
		// decide when to turn off - the end of heating season,
		// can be conservative since there is nearly no cost to
		// maintaining heating mode while not heating (keeps buffer
		// near room temperature when its warm out)
		if min_outdoor_air_temp >= zero_load_outdoor_air_temperature {
			return SystemMode{
				Off, time.Now(),
			}
		}

	case Cool:
		// decide when to turn off - the end of cooling season,
		if avg_outdoor_air_temp <= target_indoor_air_temperature {
			return SystemMode{
				Off, time.Now(),
			}
		}

	case Off:
		// if the average outdoor temperature is higher than our cooling setpoint, then
		// the room temperatures will start to rise above that setpoint, which triggers
		// cooling mode
		if (avg_outdoor_air_temp > cooling_setpoint) && RoomTooHot(avg_setpoint_err) {
			return SystemMode{
				Cool, time.Now(),
			}
		}
		// if the average outdoor temp is less than the zero-load, then room temperatures will
		// start to drop until they become too cold, which triggers heating mode
		if (avg_outdoor_air_temp < zero_load_outdoor_air_temperature) && RoomTooCold(avg_setpoint_err) {
			return SystemMode{
				Heat, time.Now(),
			}
		}
	}

	// no change
	return mode
}

// UpdateSupplyWaterTemperature calculates the ideal supply water temperature (Celsius)
func UpdateSupplyWaterTemperature(mode SystemMode, outdoor_air_temp, dewPoint float32) float32 {
	switch mode.Mode {
	case Heat:
		// linear relationship from no-load to design-load
		// re: Heating Load Line Chart
		min_heating_supply_temperature := target_indoor_air_temperature
		m := (design_supply_temperature - min_heating_supply_temperature) /
			(design_outdoor_air_temperature - zero_load_outdoor_air_temperature)
		b := design_supply_temperature - (m * design_outdoor_air_temperature)
		t := outdoor_air_temp
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
		return max(dewPoint+1.5, target_supply_temperature)

	default:
		// No supply water needed in Off mode, but return
		// a sane default anyway
		return target_indoor_air_temperature
	}
}

// UpdateCirculatorState determines whether the circulator should run.
func UpdateCirculatorState(mode SystemMode, supply_water_temperature, avg_setpoint_err float32) bool {
	switch mode.Mode {
	case Heat:
		if RoomTooHot(avg_setpoint_err) ||
			supply_water_temperature < max_indoor_air_temperature {
			return false
		}

	case Cool:
		if RoomTooCold(avg_setpoint_err) ||
			supply_water_temperature > max_indoor_air_temperature {
			return false
		}

	case Off:
		return false
	}
	return true
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
	outdoor_air_temp = -35 // Set the outdoor temperature
	outdoor_air_temp_trailing = -10
	avg_setpoint_err = -1.0
	dewPoint = 45.0 // Set the dew point

	DecisionLogic()
}
