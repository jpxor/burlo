package weather

type Forcast struct {
	Temperature         []float32
	RelHumidity         []float32
	ProbPrecipitation   []float32
	PrecipitationAmount []float32
	CloudCover          []float32
}

type Current struct {
	Temperature   float32
	RelHumidity   float32
	WindSpeed     float32
	CloudCover    float32
	Precipitation float32
	WeatherCode   int32
}

type WeatherService interface {
	CurrentConditions() (Current, error)
	Forcast24h() (Forcast, error)
}
