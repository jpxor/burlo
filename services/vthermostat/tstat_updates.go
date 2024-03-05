package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var tstat_update_queue = make(chan Thermostat)

// this is called from the mqtt message handler callback.
// that handler can't block so it spawns a goroutine to handle
// the sensor updates, and that goroutine is queued up here
// to ensure only one at a time are processed (no need to parallel)
func async_process_thermostat_update(tstat Thermostat) {
	tstat_update_queue <- tstat
}

// loop forever pulling tstats from the channel
func process_thermostat_updates() {
	defer global.waitgroup.Done()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("[process_thermostat_updates] started")
	for {
		select {
		case tstat := <-tstat_update_queue:
			sync_process_thermostat_update(tstat)

		case <-ctx.Done():
			log.Println("[process_thermostat_updates] stopped")
			return
		}
	}
}

func sync_process_thermostat_update(tstat Thermostat) {
	// the sensors don't provide dewpoint, but it is critical when cooling
	tstat.State.DewPoint = calculate_dewpoint_simple(tstat.State.Temperature, tstat.State.Humidity)

	notify_controller(tstat)
	update_tstat_history(tstat)
}

// a simple approximation, should err on the side of being
// too high, but not too low
func calculate_dewpoint_simple(temp, relH float32) float32 {
	if relH >= 50 && temp >= 25 {
		return temp - ((100 - relH) / 5)
	} else {
		return temp - ((100 - relH) / 4)
	}
}
