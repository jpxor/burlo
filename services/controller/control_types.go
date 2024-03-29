package main

import (
	"slices"
	"time"

	"golang.org/x/exp/constraints"
)

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
	if cv.Value == val {
		return
	}
	cv.Value = val
	cv.LastUpdate = time.Now()
}

func (cv *ControlValue[T]) Get() T {
	return cv.Value
}

func (cm *ControlMode[T]) Set(mode T) {
	assert(slices.Contains(cm.ValidModes, mode))
	if cm.Mode == mode {
		return
	}
	cm.Mode = mode
	cm.LastUpdate = time.Now()
}

func (cm *ControlMode[T]) Get() T {
	return cm.Mode
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
