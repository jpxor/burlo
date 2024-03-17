package main

import (
	. "burlo/services/controller/model"
	"log"
	"time"
)

type Value[T any] struct {
	Value      T
	LastUpdate time.Time
}

type Circulator struct {
	Active Value[bool]
}

type Heatpump struct {
	Mode          Value[string]
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
		log.Println("[mode] -- last update was less than 24h ago")
		return state
	}

	switch state.Heatpump.Mode.Value {
	case "heat":
		// decide when to turn off - the end of heating season,
		// can be conservative since there is nearly no cost to
		// maintaining heating mode while not heating (keeps buffer
		// near room temperature when its warm out)
		if conditions.OutdoorAir24hLow > zero_load_outdoor_air_temperature {
			log.Println("[mode] heat-->off: outdoor air above zero load temperature")
			state.Heatpump.Mode = newValue("off")
			return state
		}

	case "cool":
		// decide when to turn off - don't want this running too
		// often, so we are aggresive when turning it off
		if conditions.OutdoorAir24hAvg < cooling_mode_cutoff &&
			!RoomTooHot(conditions.SetpointError) {
			log.Println("[mode] cool->off: mean outdoor air below cooling trigger temperature")
			state.Heatpump.Mode = newValue("off")
			return state
		}

	case "off":
		// Note: there is no setpoint when system is off! Don't know if we need
		// the heating or cooling setpoint at the time setpoint error is calculated

		// if the average outdoor temperature is higher than our cooling setpoint, then
		// the room temperatures will start to rise, which triggers cooling mode
		if conditions.OutdoorAir24hAvg > design_indoor_air_temperature { // TODO: use config.CoolingTriggerTemperature
			log.Println("[mode] off->cool: mean outdoor air above cooling trigger temperature and room too hot")
			state.Heatpump.Mode = newValue("cool")
			return state
		}
		// if the average outdoor temp is less than the zero-load, then room temperatures will
		// start to drop, which triggers heating mode
		if conditions.OutdoorAir24hAvg < zero_load_outdoor_air_temperature {
			log.Println("[mode] off->heat: mean outdoor air below zero load temperature and room too cold")
			state.Heatpump.Mode = newValue("heat")
			return state
		}
	}

	// no change
	log.Println("[mode] -- no change, state =", state.Heatpump.Mode.Value)
	return state
}

func update_supply_temp(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	switch state.Heatpump.Mode.Value {
	case "heat":
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
		return state

	case "cool":
		// cooling temperature is set to ensure the floors don't
		// get uncomfortably cool. We can go lower during night,
		// But it must remain above dewpoint at all times!
		target_supply_temperature := comfort_cooling_supply_temperature
		if NightCoolingBoost() {
			target_supply_temperature = min_cooling_supply_temperature
		}
		state.TsTemperature = newValue(max(conditions.DewPoint+1.5, target_supply_temperature))
		log.Println("[Tsupply] cooling", state.TsTemperature.Value, "C supply temperature")
		return state

	default:
		// No supply water needed in Off mode, but return
		// a sane default anyway
		state.TsTemperature = newValue(design_indoor_air_temperature)
		state.TsCorrection = newValue(float32(0))
		return state
	}
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
	case "heat":
		if RoomTooHot(conditions.SetpointError) ||
			state.TsTemperature.Value < conditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: room too hot or Ts too low")
			state.Circulator.Active = newValue(false)
			return state
		}

	case "cool":
		if RoomTooCold(conditions.SetpointError) ||
			state.TsTemperature.Value > conditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: room too cold or Ts too high")
			state.Circulator.Active = newValue(false)
			return state
		}

	case "off":
		state.Circulator.Active = newValue(false)
		return state
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
