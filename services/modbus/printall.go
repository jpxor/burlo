package main

import (
	"log"
	"os"
	"time"

	"github.com/simonvetter/modbus"
)

func printRegisters(cfg Config) {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     cfg.DeviceURI,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.Println("Error creating Modbus client:", err)
		return
	}

	err = client.Open()
	for err != nil {
		log.Println("faild to open", err)
		os.Exit(1)
	}

	defer client.Close()
	client.SetUnitId(cfg.DeviceID)

	for _, register := range cfg.Register {
		var reg16 uint16
		retry := 3
		for {
			reg16, err = client.ReadRegister(register.Address, modbus.HOLDING_REGISTER)
			if err != nil {
				if retry > 0 {
					retry -= 1
					continue
				}
				log.Println(register.Address, "ERR", err)
			}
			break
		}
		switch register.Type {
		case BOOL:
			var value bool
			value = reg16 > 0
			log.Printf("%v %v --> %v %v\n", register.Address, register.Name, value, register.Units)

		case INT16:
			var value float32
			value = float32(int16(reg16)) * register.Factor
			log.Printf("%v %v --> %v %v\n", register.Address, register.Name, value, register.Units)

		case UINT16:
			var value float32
			value = float32(reg16) * register.Factor
			log.Printf("%v %v --> %v %v\n", register.Address, register.Name, value, register.Units)
		}

	}
}
