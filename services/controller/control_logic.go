package main

import (
	. "burlo/services/controller/model"
	"fmt"
	"log"
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
var cooling_mode_high_temp_trigger float32 = 28

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

func clamp(minv, v, maxv float32) float32 {
	v = min(v, maxv)
	v = max(v, minv)
	return v
}

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

func applyV2(state SystemStateV2) {
	log.Println("[controller] applying state")
}
