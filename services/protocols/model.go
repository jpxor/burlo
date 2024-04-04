package protocol

type OutdoorConditions struct {
	OutdoorAirTemp    float32
	OutdoorAir24hLow  float32
	OutdoorAir24hAvg  float32
	OutdoorAir24hHigh float32
}

type Thermostat struct {
	ID           string
	Name         string
	Temperature  float32
	Humidity     float32
	DewPoint     float32
	HeatSetpoint float32
	CoolSetpoint float32
}

type SensorData struct {
	Battery     int32
	LinkQuality int32
	Temperature float32
	Humidity    float32
	DewPoint    float32
}

type PhidgetDO struct {
	Name    string
	Channel int32
	HubPort int32
	Output  bool
}

type ModbusReg struct {
	Name     string
	Register int32
	Value    float32
}
