package main

import (
	. "burlo/services/controller/model"
	"log"
	"math"
	"time"
)

type Value[T any] struct {
	Value      T
	LastUpdate time.Time
}

type Circulator struct {
	Active Value[bool]
}

type HPMode string

const (
	HEAT HPMode = "Heat"
	COOL HPMode = "Cool"
)

type Heatpump struct {
	Mode          Value[HPMode]
	TsTemperature Value[float32]
	TsCorrection  Value[float32]
}

type SystemStateV2 struct {
	Circulator
	Heatpump
}

func initValue[T any](val T) Value[T] {
	return Value[T]{val, time.Time{}}
}

func newValue[T any](val T) Value[T] {
	return Value[T]{val, time.Now()}
}

func isInitialized(conditions ControlConditions) bool {
	return !conditions.IndoorConditions.LastUpdate.IsZero() &&
		!conditions.OutdoorConditions.LastUpdate.IsZero()
}

func system_update(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	if !isInitialized(conditions) {
		return state
	}
	log.Println("[controller] system update")
	state = update_mode(state, conditions)
	state = update_supply_temp(state, conditions)
	state = update_circulator(state, conditions)

	applyV2(state)
	update_history(state, conditions)

	return state
}

func update_mode(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	// debounce, don't let the mode switch too often
	if time.Since(state.Heatpump.Mode.LastUpdate) < 24*time.Hour {
		return state
	}
	switch state.Heatpump.Mode.Value {
	case HEAT:
		// decide when to switch from heating to cooling
		if conditions.OutdoorAir24hLow > zero_load_outdoor_air_temperature &&
			conditions.OutdoorAir24hAvg > design_indoor_air_temperature &&
			conditions.OutdoorAir24hHigh > cooling_mode_high_temp_trigger {
			log.Println("[mode] heat --> cool")
			state.Heatpump.Mode = newValue(COOL)
			return state
		}
	case COOL:
		// decide when to switch from cooling to heating
		if conditions.OutdoorAir24hLow < zero_load_outdoor_air_temperature &&
			conditions.OutdoorAir24hAvg < design_indoor_air_temperature {
			log.Println("[mode] cool --> heat")
			state.Heatpump.Mode = newValue(HEAT)
			return state
		}
	}
	// no change
	return state
}

func update_supply_temp(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	switch state.Heatpump.Mode.Value {
	case HEAT:
		// linear relationship from no-load to design-load
		// re: Heating Load Line Chart
		min_heating_supply_temperature := design_indoor_air_temperature
		m := (design_supply_temperature - min_heating_supply_temperature) /
			(design_outdoor_air_temperature - zero_load_outdoor_air_temperature)
		b := design_supply_temperature - (m * design_outdoor_air_temperature)
		t := conditions.OutdoorAirTemp
		target_supply_temperature := m*t + b

		// correction with delay
		if state.Heatpump.TsCorrection.LastUpdate.IsZero() {
			state.Heatpump.TsCorrection.LastUpdate = time.Now()
		}
		if time.Since(state.Heatpump.TsCorrection.LastUpdate) > 15*time.Minute {
			if RoomTooHot(conditions.SetpointError) {
				log.Println("[Tsupply-correction] -1: room too hot")
				state.TsCorrection.Value -= 1
				state.TsCorrection.LastUpdate = time.Now()
			}
			if RoomTooCold(conditions.SetpointError) &&
				target_supply_temperature < max_supply_temperature {
				log.Println("[Tsupply-correction] +1: room too cold")
				state.TsCorrection.Value += 1
				state.TsCorrection.LastUpdate = time.Now()
			}
		}
		target_supply_temperature += state.TsCorrection.Value
		state.TsTemperature = newValue(min(max_supply_temperature, max(min_heating_supply_temperature, target_supply_temperature)))
		log.Println("[Tsupply] heating", state.TsTemperature.Value, "C supply temperature")

	case COOL:
		// cooling temperature is set to ensure the floors don't
		// get uncomfortably cool. We can go lower during night,
		// But it must remain above dewpoint at all times!
		target_supply_temperature := comfort_cooling_supply_temperature
		if NightCoolingBoost() {
			target_supply_temperature = min_cooling_supply_temperature
		}
		state.TsTemperature = newValue(max(conditions.DewPoint+1.5, target_supply_temperature))
		log.Println("[Tsupply] cooling", state.TsTemperature.Value, "C supply temperature")
	}
	return state
}

func update_circulator(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	// debounce, don't let the circulator switch too often
	if state.Circulator.Active.Value {
		// run for at least 15 mins
		if time.Since(state.Circulator.Active.LastUpdate) < 15*time.Minute {
			return state
		}
	} else {
		// stay off for at least 1 min
		if time.Since(state.Circulator.Active.LastUpdate) < 1*time.Minute {
			return state
		}
	}

	switch state.Heatpump.Mode.Value {
	case HEAT:
		if RoomTooHot(conditions.SetpointError) ||
			state.TsTemperature.Value < conditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: room too hot or Ts too low")
			state.Circulator.Active = newValue(false)
			return state
		}

	case COOL:
		if RoomTooCold(conditions.SetpointError) ||
			state.TsTemperature.Value > conditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: room too cold or Ts too high")
			state.Circulator.Active = newValue(false)
			return state
		}
	}

	// no change if already running
	if state.Circulator.Active.Value {
		return state
	}

	// start running if its not
	state.Circulator.Active = newValue(true)
	log.Println("[cirlculator] on")
	return state
}

func applyV2(state SystemStateV2) {
	log.Println("[controller] applying state")
}

func clamp(minv, v, maxv float32) float32 {
	v = min(v, maxv)
	v = max(v, minv)
	return v
}
