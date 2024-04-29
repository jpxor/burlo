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
	Name    string `json:"name"`
	Channel int32  `json:"channel"`
	HubPort int32  `json:"hub_port"`
	Output  bool   `json:"target_state"`
}

type PhidgetVO struct {
	Name    string  `json:"name"`
	Channel int32   `json:"channel"`
	HubPort int32   `json:"hub_port"`
	Output  float32 `json:"target_state"`
}
