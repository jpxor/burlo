package main

import (
	"slices"
	"time"

	"golang.org/x/exp/constraints"
)

type Mode string

const ( // circulator modes
	ON  Mode = "ON"
	OFF      = "OFF"
)

const ( // heatpump modes
	HEAT Mode = "Heat"
	COOL      = "Cool"
)

type Controls struct {
	Circulator ControlMode[Mode]
	Heatpump   ControlMode[Mode]
	SupplyTemp ControlValue[float32]
}

type ControlValue[T constraints.Ordered] struct {
	Value      T
	Min        T
	Max        T
	LastUpdate time.Time
}

type ControlMode[T comparable] struct {
	Mode       T
	ValidModes []T
	LastUpdate time.Time
}

func (cv *ControlValue[T]) Set(val T) {
	val = clamp(cv.Min, val, cv.Max)
	if cv.Value == val && !cv.LastUpdate.IsZero() {
		return
	}
	cv.Value = val
	cv.LastUpdate = time.Now()
}

func (cm *ControlMode[T]) Set(mode T) {
	assert(slices.Contains(cm.ValidModes, mode))
	if cm.Mode == mode && !cm.LastUpdate.IsZero() {
		return
	}
	cm.Mode = mode
	cm.LastUpdate = time.Now()
}

func assert(pass bool) {
	if !pass {
		panic("failed assert")
	}
}

func clamp[T constraints.Ordered](minv, v, maxv T) T {
	v = min(v, maxv)
	v = max(v, minv)
	return v
}
