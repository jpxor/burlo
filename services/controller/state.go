package main

import (
	. "burlo/services/controller/model"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
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

func isInitialized(conditions ControlConditions) bool {
	return !conditions.IndoorConditions.LastUpdate.IsZero() &&
		!conditions.OutdoorConditions.LastUpdate.IsZero()
}

func control_loop(indoorChan chan IndoorConditions, outdoorChan chan OutdoorConditions) {

	system_state := SystemStateV2{
		Circulator{
			initValue(false),
		},
		Heatpump{
			Mode:          initValue("off"),
			TsTemperature: initValue(float32(20)),
			TsCorrection:  initValue(float32(0)),
		},
	}

	system_conditions := ControlConditions{
		IndoorConditions{},
		OutdoorConditions{},
	}

	log.Println("[controller] started")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case idc := <-indoorChan:
			system_conditions.IndoorConditions = idc
			system_conditions.IndoorConditions.LastUpdate = time.Now()

		case odc := <-outdoorChan:
			system_conditions.OutdoorConditions = odc
			system_conditions.OutdoorConditions.LastUpdate = time.Now()

		case <-ctx.Done():
			log.Println("[controller] stopped")
			return
		}
		if !isInitialized(system_conditions) {
			continue
		}
		system_state = update(system_state, system_conditions)
		applyV2(system_state)

		update_history(system_state, system_conditions)
	}
}

func update(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	state = update_mode(state, conditions)
	state = update_supply_temp(state, conditions)
	state = update_circulator(state, conditions)
	return state
}

func update_mode(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	return state
}

func update_supply_temp(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	return state
}

func update_circulator(state SystemStateV2, conditions ControlConditions) SystemStateV2 {
	return state
}

func applyV2(state SystemStateV2) {
	//
}
