package main

import (
	"burlo/pkg/models/controller"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func update(tstat controller.Thermostat) {
	mutex.Lock()
	defer mutex.Unlock()

	// default values for the thermostat
	tstat.HeatSetpoint = 21
	tstat.CoolSetpoint = 24

	// id needs to be safe to use in URL paths
	// as well as in css class names
	tstat.ID = safeID(tstat.ID)

	// match with existing tstat
	existing, ok := thermostats[tstat.ID]
	if ok {
		tstat.Name = existing.Name
		tstat.HeatSetpoint = existing.HeatSetpoint
		tstat.CoolSetpoint = existing.CoolSetpoint
	}
	tstat.Time = time.Now()
	tstat.Dewpoint = calculate_dewpoint_simple(tstat.Temperature, tstat.Humidity)

	thermostats[tstat.ID] = tstat
	publish(tstat)
}

// writes to mqtt for other services to consume
func publish(tstat controller.Thermostat) {
	const RETAIN = true
	topic := fmt.Sprintf("controller/thermostats/%s", tstat.ID)
	publisher.Publish(RETAIN, topic, tstat)
}

func safeID(id string) string {
	id = url.PathEscape(id)
	return strings.ReplaceAll(id, "%", "_")
}

// a simple approximation, should err on the side of
// being too high, never too low. Temperature must be
// in celcius. Accureate to within 1 degC when RelH > 50%
func calculate_dewpoint_simple(temp, relH float32) float32 {
	if relH >= 50 && temp >= 25 {
		return temp - ((100 - relH) / 5)
	} else {
		return temp - ((100 - relH) / 4)
	}
}
