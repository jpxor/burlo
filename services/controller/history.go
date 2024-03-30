package main

import (
	"burlo/pkg/lockbox"
	"time"
)

type History struct {
	Controls
	Conditions
	Time time.Time
}

var history = lockbox.New([]History{})

// keeps all data points within the last 48 hours
func update_history(controls Controls, conditions Conditions) {
	hlist, lbk := history.Take()
	hlist = append(hlist, History{
		controls, conditions, time.Now(),
	})
	for i, h := range hlist {
		if time.Since(h.Time) <= 48*time.Hour {
			hlist = hlist[i:]
			break
		}
	}
	history.Put(hlist, lbk)
}
