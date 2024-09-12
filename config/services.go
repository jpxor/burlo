package config

import (
	"log"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

type ServiceConf struct {
	Units                string               `toml:"units"`
	ServiceHTTPAddresses ServiceHTTPAddresses `toml:"service_http_addresses"`
	Dx2WModbus           Dx2WModbus           `toml:"dx2w_modbus"`
	Location             Location             `toml:"location"`
	Thermostat           Thermostat           `toml:"thermostat"`
	Controller           Controller           `toml:"controller"`
	Mqtt                 Mqtt                 `toml:"mqtt"`
}
type ServiceHTTPAddresses struct {
	Dx2Wlogger string `toml:"dx2wlogger"`
	Controller string `toml:"controller"`
	Thermostat string `toml:"thermostat"`
	Mqttserver string `toml:"mqttserver"`
	Actuators  string `toml:"actuators"`
	Dashboard  string `toml:"dashboard"`
	NtfyServer string `toml:"ntfyserver"`
}
type Dx2WModbus struct {
	TCPAddress string `toml:"tcp_address"`
	DeviceID   uint8  `toml:"device_id"`
}
type Location struct {
	Latitude  string `toml:"latitude"`
	Longitude string `toml:"longitude"`
}
type Mqtt struct {
	Address string `toml:"address"`
	Prefix  string `toml:"prefix"`
	User    string `toml:"user"`
	Pass    string `toml:"pass"`
}
type Thermostat struct {
	Mqtt Mqtt `toml:"mqtt"`
}
type RadiantCooling struct {
	Enabled           bool `toml:"enabled"`
	OvernightBoost    bool `toml:"overnight_boost"`
	SupplyTemperature int  `toml:"supply_temperature"`
}
type Circulator struct {
	Hubport int    `toml:"hubport"`
	Channel int    `toml:"channel"`
	Type    string `toml:"type"`
}
type Hpmode struct {
	Hubport int    `toml:"hubport"`
	Channel int    `toml:"channel"`
	Type    string `toml:"type"`
}
type Dewpoint struct {
	Hubport int    `toml:"hubport"`
	Channel int    `toml:"channel"`
	Type    string `toml:"type"`
}
type Phidgets struct {
	Circulator Circulator `toml:"circulator"`
	Hpmode     Hpmode     `toml:"hpmode"`
	Dewpoint   Dewpoint   `toml:"dewpoint"`
}
type Controller struct {
	RadiantCooling RadiantCooling `toml:"radiant_cooling"`
	Phidgets       Phidgets       `toml:"phidgets"`
}

func LoadV2(filepath string) ServiceConf {
	cfg_str, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalln(err)
	}
	var cfg ServiceConf
	err = toml.Unmarshal(cfg_str, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}

func GetPort(addr string) string {
	splits := strings.Split(addr, ":")
	if len(splits) == 2 {
		return splits[1]
	}
	return ""
}
