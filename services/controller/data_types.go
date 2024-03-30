package main

import (
	protocol "burlo/services/protocols"
	"time"
)

type Thermostat struct {
	ID           string
	HeatSetpoint float32
	CoolSetpoint float32
	Temperature  float32
	DewPoint     float32
	LastUpdate   time.Time
}

func (self *Thermostat) From(t protocol.Thermostat) {
	self.ID = t.ID
	self.HeatSetpoint = t.HeatSetpoint
	self.CoolSetpoint = t.CoolSetpoint
	self.Temperature = t.Temperature
	self.DewPoint = t.DewPoint
	self.LastUpdate = time.Now()
}

type Conditions struct {
	IndoorConditions
	OutdoorConditions
}

type IndoorConditions struct {
	HeatSetpointError float32
	CoolSetpointError float32
	DewPoint          float32
	IndoorAirTempMax  float32
	LastUpdate        time.Time
}

func (self *IndoorConditions) From(tstats map[string]Thermostat) {
	for _, t := range tstats {
		self.IndoorAirTempMax = max(self.IndoorAirTempMax, t.Temperature)
		self.DewPoint = max(self.DewPoint, t.DewPoint)

		self.HeatSetpointError += (t.Temperature - t.HeatSetpoint)
		self.CoolSetpointError += (t.Temperature - t.CoolSetpoint)
	}
	self.HeatSetpointError /= float32(len(global.thermostats))
	self.CoolSetpointError /= float32(len(global.thermostats))
	self.LastUpdate = time.Now()
}

type OutdoorConditions struct {
	OutdoorAirTemp    float32
	OutdoorAir24hLow  float32
	OutdoorAir24hAvg  float32
	OutdoorAir24hHigh float32
	LastUpdate        time.Time
}

func (self *OutdoorConditions) From(o protocol.OutdoorConditions) {
	self.OutdoorAirTemp = o.OutdoorAirTemp
	self.OutdoorAir24hLow = o.OutdoorAir24hLow
	self.OutdoorAir24hAvg = o.OutdoorAir24hAvg
	self.OutdoorAir24hHigh = o.OutdoorAir24hHigh
	self.LastUpdate = time.Now()
}
