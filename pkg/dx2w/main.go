package dx2w

import (
	_ "embed"
)

//go:embed dx2w-modbus.toml
var modbusConf []byte

var cfg = ParseConfig(modbusConf)
