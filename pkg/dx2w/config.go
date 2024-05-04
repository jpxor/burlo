package dx2w

import (
	"cmp"
	"log"
	"slices"

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

func ParseConfig(rawconf []byte) Config {
	var cfg Config
	err := toml.Unmarshal(rawconf, &cfg)
	if err != nil {
		log.Fatalln(err)
	}
	slices.SortFunc(cfg.Register, func(a, b Register) int {
		return cmp.Compare(a.Address, b.Address)
	})
	return cfg
}
