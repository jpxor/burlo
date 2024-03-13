package config

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type Configuration struct {
	Units    string `toml:"units"`
	Services struct {
		Controller   string `toml:"controller"`
		Vthermostats string `toml:"vthermostats"`
		MqttServer   string `toml:"mqtt_server"`
		Actuators    string `toml:"actuators"`
	} `toml:"services"`
	Weather struct {
		Latitude  string `toml:"latitude"`
		Longitude string `toml:"longitude"`
	} `toml:"weather"`
	Controller struct {
		Cooling struct {
			Enabled                   bool `toml:"enabled"`
			OvernightBoostEnabled     bool `toml:"overnight_boost_enabled"`
			OvernightBoostTemperature int  `toml:"overnight_boost_temperature"`
			CoolingSupplyTemperature  int  `toml:"cooling_supply_temperature"`
			CoolingTriggerTemperature int  `toml:"cooling_trigger_temperature"`
		} `toml:"cooling"`
		Heating struct {
			MaxSupplyTemperature            float64 `toml:"max_supply_temperature"`
			DesignLoadOutdoorAirTemperature int     `toml:"design_load_outdoor_air_temperature"`
			DesignLoadSupplyTemperature     float64 `toml:"design_load_supply_temperature"`
			ZeroLoadOutdoorAirTemperature   int     `toml:"zero_load_outdoor_air_temperature"`
			ZeroLoadSupplyTemperature       int     `toml:"zero_load_supply_temperature"`
		} `toml:"heating"`
	} `toml:"controller"`
	Thermostats struct {
		Primary string `toml:"primary"`
	} `toml:"thermostats"`
	Actuators struct {
		Circulator struct {
			Hubport int    `toml:"hubport"`
			Channel int    `toml:"channel"`
			Type    string `toml:"type"`
		} `toml:"circulator"`
	} `toml:"actuators"`
	Mqtt struct {
		Prefix string `toml:"prefix"`
		User   string `toml:"user"`
		Pass   string `toml:"pass"`
	} `toml:"mqtt"`
}

func Load(filepath string) Configuration {
	cfg_str, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalln(err)
	}
	var cfg Configuration
	err = toml.Unmarshal(cfg_str, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}
