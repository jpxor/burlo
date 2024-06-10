package dx2w

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

//go:embed dx2w-modbus.toml
var modbusConf []byte

var globalRegisterConfig = parseConfig(modbusConf)

type TCPDevice struct {
	Url string
	Id  uint8
}

type TCPClient struct {
	device TCPDevice
	config Config
}

type Value struct {
	Float32   float32
	Uint16    uint16
	Bool      bool
	Type      DataType
	Units     string
	Timestamp time.Time
}

func New(device TCPDevice) *TCPClient {
	client := &TCPClient{
		device: device,
		config: globalRegisterConfig,
	}
	return client
}

func NewWithFields(device TCPDevice, fields []string) *TCPClient {
	client := New(device)
	client.config = client.config.withFields(fields)
	return client
}

func (c TCPClient) PrintFields() {
	for _, reg := range c.config.Register {
		fmt.Println(reg.Name)
	}
}

func (c TCPClient) ReadAll() map[string]Value {
	regmap := make(map[string]Value)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     c.device.Url,
		Timeout: 4 * time.Second,
	})
	if err != nil {
		fmt.Println("Error creating Modbus client:", err)
		return regmap
	}

	err = client.Open()
	for err != nil {
		fmt.Println("Error opening Modbus client:", err)
		return regmap
	}

	defer client.Close()
	client.SetUnitId(c.device.Id)

	now := time.Now()
	nretries := 0

	nread := 0
	for nread < len(c.config.Register) {

		// registers that have not yet been read
		registers := c.config.Register[nread:]

		// will read up to 16 registers per request
		firstAddr := registers[0].Address
		lastAddr := firstAddr + 16
		count := 0

		for _, reg := range registers {
			if reg.Address >= lastAddr {
				break
			}
			count += 1
		}

		// the actual last address
		lastAddr = registers[count-1].Address
		nregisters := 1 + lastAddr - firstAddr

		rawvals, err := client.ReadRegisters(firstAddr, nregisters, modbus.HOLDING_REGISTER)

		if err != nil {
			if nretries < 3 {
				time.Sleep(100 * time.Millisecond)
				nretries += 1
				continue
			}
			fmt.Printf("Failed to read modbus registers %v - %v: %v", firstAddr, lastAddr, err)
			nretries = 0

		} else {
			for _, reg := range registers[:count] {
				i := reg.Address - firstAddr
				regmap[reg.Name] = Value{
					Uint16:    rawvals[i],
					Float32:   float32(int16(rawvals[i])) * reg.Factor,
					Bool:      asBool(rawvals[i]),
					Type:      reg.Type,
					Units:     reg.Units,
					Timestamp: now,
				}
			}
		}
		nread += count
	}
	return regmap
}

func asBool(val uint16) bool {
	if val != 0 {
		return true
	}
	return false
}
