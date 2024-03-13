package model

type Forcast struct {
	Temperatures []float32
}

type Conditions struct {
	Temperature float32
	RelHumidity float32
	WindSpeed   float32
	CloudCover  float32
	WeatherCode int32
}
