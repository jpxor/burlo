package main

import (
	protocol "burlo/services/protocols"
	"fmt"
	"log"
	"time"
)

func update_outdoor_conditions(odc OutdoorConditions) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	global.OutdoorConditions = odc
	update_controls_locked()
}

func update_indoor_conditions(tstat Thermostat) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	// invalidate old data before adding new
	for id, t := range global.thermostats {
		if time.Since(t.LastUpdate) > 24*time.Hour {
			delete(global.thermostats, id)
		}
	}
	global.thermostats[tstat.ID] = tstat

	global.IndoorConditions.From(global.thermostats)
	update_controls_locked()
}

// global mutex lock must be held when calling
// these update functions
func update_controls_locked() {

	// can't update controls until both outdoor and
	// indoor conditions get at least 1 update
	if global.OutdoorConditions.LastUpdate.IsZero() ||
		global.IndoorConditions.LastUpdate.IsZero() {
		return
	}

	log.Println("[controller] update_controls")
	update_mode()
	update_supply_temp()
	update_circulator()

	applyV2(global.Controls)
	update_history(global.Controls, global.Conditions)
}

func update_mode() {
	if !cooling_enabled {
		global.Heatpump.Set(HEAT)
		return
	}
	// debounce, don't let the mode switch too often
	if time.Since(global.Heatpump.LastUpdate) < 24*time.Hour {
		return
	}
	switch global.Heatpump.Mode {
	case HEAT:
		// decide when to switch from heating to cooling
		if global.OutdoorAir24hLow > zero_load_outdoor_air_temperature &&
			global.OutdoorAir24hAvg > design_indoor_air_temperature &&
			global.OutdoorAir24hHigh > cooling_mode_high_temp_trigger {
			log.Println("[mode] heat --> cool")
			global.Heatpump.Set(COOL)
		}

	case COOL:
		// decide when to switch from cooling to heating
		if global.OutdoorAir24hLow < zero_load_outdoor_air_temperature &&
			global.OutdoorAir24hAvg < design_indoor_air_temperature {
			log.Println("[mode] cool --> heat")
			global.Heatpump.Set(HEAT)
		}
	}
}

func update_supply_temp() {
	switch global.Heatpump.Mode {
	case HEAT:
		// linear relationship from no-load to design-load
		// re: Heating Load Line Chart
		min_heating_supply_temperature := design_indoor_air_temperature
		m := (design_supply_temperature - min_heating_supply_temperature) /
			(design_outdoor_air_temperature - zero_load_outdoor_air_temperature)
		b := design_supply_temperature - (m * design_outdoor_air_temperature)
		t := global.OutdoorAirTemp
		target_supply_temperature := m*t + b

		// correction with delay
		// if state.Heatpump.TsCorrection.LastUpdate.IsZero() {
		// 	state.Heatpump.TsCorrection.LastUpdate = time.Now()
		// }
		// if time.Since(state.Heatpump.TsCorrection.LastUpdate) > 15*time.Minute {
		// 	if RoomTooHot(conditions.SetpointError) {
		// 		log.Println("[Tsupply-correction] -1: room too hot")
		// 		state.TsCorrection.Value -= 1
		// 		state.TsCorrection.LastUpdate = time.Now()
		// 	}
		// 	if RoomTooCold(conditions.SetpointError) &&
		// 		target_supply_temperature < max_supply_temperature {
		// 		log.Println("[Tsupply-correction] +1: room too cold")
		// 		state.TsCorrection.Value += 1
		// 		state.TsCorrection.LastUpdate = time.Now()
		// 	}
		// 	if math.Abs(float64(state.TsCorrection.Value)) > 5 {
		// 		log.Println("[WARN] supply temperature correction too large")
		// 		state.TsCorrection.Value = clamp(-5, state.TsCorrection.Value, 5)
		// 	}
		// }
		// target_supply_temperature += state.TsCorrection.Value

		global.SupplyTemp.Min = min_heating_supply_temperature
		global.SupplyTemp.Max = max_supply_temperature
		global.SupplyTemp.Set(target_supply_temperature)
		log.Println("[Tsupply] heating", global.SupplyTemp.Value, "C supply temperature")

	case COOL:
		// cooling temperature is set to ensure the floors don't
		// get uncomfortably cool. We can go lower during night,
		// But it must remain above dewpoint at all times!
		target_supply_temperature := comfort_cooling_supply_temperature
		if NightCoolingBoost() {
			target_supply_temperature = min_cooling_supply_temperature
		}
		global.SupplyTemp.Min = global.DewPoint + 1.5
		global.SupplyTemp.Max = max_supply_temperature
		global.SupplyTemp.Set(target_supply_temperature)
		log.Println("[Tsupply] cooling", global.SupplyTemp.Value, "C supply temperature")
	}
}

func update_circulator() {
	// debounce, don't let the circulator switch too often
	// run for at least 15 mins
	// stay off for at least 1 min
	if global.Circulator.Mode == ON {
		if time.Since(global.Circulator.LastUpdate) < 15*time.Minute {
			return
		}
	} else {
		if time.Since(global.Circulator.LastUpdate) < 1*time.Minute {
			return
		}
	}

	switch global.Heatpump.Mode {
	case HEAT:
		if RoomTooHot(global.HeatSetpointError) {
			log.Println("[cirlculator] off: room too hot")
			global.Circulator.Set(OFF)
			return
		}
		if !RoomTooCold(global.HeatSetpointError) &&
			global.Conditions.OutdoorAirTemp > global.IndoorConditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: outdoor temp warmer than indoor (no heat loss)")
			global.Circulator.Set(OFF)
			return
		}

	case COOL:
		if RoomTooCold(global.CoolSetpointError) {
			log.Println("[cirlculator] off: room too cold")
			global.Circulator.Set(OFF)
			return
		}
		if !RoomTooHot(global.CoolSetpointError) &&
			global.Conditions.OutdoorAirTemp < global.IndoorConditions.IndoorAirTempMax {
			log.Println("[cirlculator] off: outdoor temp cooler than indoor (no heat gain)")
			global.Circulator.Set(OFF)
			return
		}
	}

	if global.Circulator.Mode != ON {
		log.Println("[cirlculator] on")
	}
	global.Circulator.Set(ON)
}

// RoomTooCold is a helper that returns true if the
// room temperature falls below the target (with some margin)
func RoomTooCold(setpoint_error float32) bool {
	return setpoint_error <= -0.5 // example: {target=20, too_cold=19.5}
}

// RoomTooHot is a helper that returns true if the
// room temperature rises above the target (with some margin)
func RoomTooHot(setpoint_error float32) bool {
	return setpoint_error >= 0.5 // example: {target=20, too_hot=20.5}
}

func NightCoolingBoost() bool {
	if !cooling_overnight_boost_enabled {
		return false
	}
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

func applyV2(controls Controls) {
	log.Println("[controller] applying state")
	set_digital_out(protocol.PhidgetDO{
		Name:    "CoolingMode",
		HubPort: 0,
		Channel: 1,
		Output:  controls.Heatpump.Mode == COOL,
	})
	set_digital_out(protocol.PhidgetDO{
		Name:    "ZoneCirculator",
		HubPort: 0,
		Channel: 0,
		Output:  controls.Circulator.Mode == ON,
	})
	set_voltage_out(protocol.PhidgetVO{
		Name:    "Dewpoint",
		HubPort: 1,
		Channel: 0,
		Output:  DewpointToVoltage(global.IndoorConditions.DewPoint),
	})
}

func DewpointToVoltage(temperature float32) float32 {
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
