package main

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type DataType string

var INT16 DataType = "INT16"
var UINT16 DataType = "UINT16"
var BOOL DataType = "BOOL"

type Register struct {
	Name     string
	Address  uint16
	Factor   float32
	Writable bool
	Type     DataType
	Units    string
}

type Config struct {
	DeviceURI string
	DeviceID  uint8
	Register  []Register
}

func LoadConfig(filepath string) Config {
	cfg_str, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalln(err)
	}
	var cfg Config
	err = toml.Unmarshal(cfg_str, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}
