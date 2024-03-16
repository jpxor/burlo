package main

import (
	"burlo/pkg/lockbox"
	. "burlo/services/controller/model"
	"time"
)

type History struct {
	SystemStateV2
	ControlConditions
	Time time.Time
}

var history = lockbox.New([]History{})

// keeps all data points within the last 48 hours
func update_history(state SystemStateV2, conditions ControlConditions) {
	hlist, lbk := history.Take()
	hlist = append(hlist, History{
		state, conditions, time.Now(),
	})
	for i, h := range hlist {
		if time.Since(h.Time) <= 48*time.Hour {
			hlist = hlist[i:]
			break
		}
	}
	history.Put(hlist, lbk)
}
