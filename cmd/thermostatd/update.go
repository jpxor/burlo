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

	// id needs to be safe to use in URL paths
	// as well as in css class names
	tstat.ID = safeID(tstat.ID)

	tstat.Time = time.Now()
	tstat.Dewpoint = calculate_dewpoint_simple(tstat.Temperature, tstat.Humidity)

	// humiditstats don't need temperature setpoints
	// and friendly/customizable names
	if tstat.DewpointOnly {
		publishHumidistat(tstat)
		return
	}

	// default values for the thermostat
	tstat.HeatSetpoint = 20
	tstat.CoolSetpoint = 24

	// match with existing tstat
	existing, ok := thermostats[tstat.ID]
	if ok {
		tstat.Name = existing.Name
		tstat.HeatSetpoint = existing.HeatSetpoint
		tstat.CoolSetpoint = existing.CoolSetpoint
	}
	thermostats[tstat.ID] = tstat
	publishThermostat(tstat)
}

// writes to mqtt for other services to consume
func publishThermostat(tstat controller.Thermostat) {
	const RETAIN = true
	topic := fmt.Sprintf("controller/thermostats/%s", tstat.ID)
	publisher.Publish(RETAIN, topic, tstat)
}

func publishHumidistat(tstat controller.Thermostat) {
	const RETAIN = true
	topic := fmt.Sprintf("controller/humidistat/%s", tstat.ID)
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
