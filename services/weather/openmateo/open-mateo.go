package openmateo

import (
	"burlo/pkg/timezone"
	weather "burlo/services/weather/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type TemperatureForcastResp struct {
	Timezone    string `json:"timezone"`
	HourlyUnits struct {
		Time        string `json:"time"`
		Temperature string `json:"temperature_2m"`
	} `json:"hourly_units"`
	Hourly struct {
		Time         []string  `json:"time"`
		Temperatures []float32 `json:"temperature_2m"`
	} `json:"hourly"`
}

type WeatherResp struct {
	Timezone     string  `json:"timezone"`
	Elevation    float32 `json:"elevation"`
	CurrentUnits struct {
		Time             string `json:"time"`
		Interval         string `json:"interval"`
		Temperature      string `json:"temperature_2m"`
		RelativeHumidity string `json:"relative_humidity_2m"`
		WeatherCode      string `json:"weather_code"`
		CloudCover       string `json:"cloud_cover"`
		WindSpeed        string `json:"wind_speed_10m"`
	} `json:"current_units"`
	Current struct {
		Time             string  `json:"time"`
		Interval         int     `json:"interval"`
		Temperature      float32 `json:"temperature_2m"`
		RelativeHumidity float32 `json:"relative_humidity_2m"`
		WeatherCode      int32   `json:"weather_code"`
		CloudCover       float32 `json:"cloud_cover"`
		WindSpeed        float32 `json:"wind_speed_10m"`
	} `json:"current"`
}

type OpenMeteoService struct {
	forcast            string
	current_conditions string
}

func New(lat, long string) (*OpenMeteoService, error) {
	tzname, err := timezone.FromCoordinates(lat, long)
	if err != nil {
		return nil, err
	}
	return &OpenMeteoService{
		forcast: fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&hourly=temperature_2m&timezone=%s&forecast_days=3",
			url.QueryEscape(lat), url.QueryEscape(long), url.QueryEscape(tzname)),
		current_conditions: fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current=temperature_2m,relative_humidity_2m,weather_code,cloud_cover,wind_speed_10m&timezone=%s",
			url.QueryEscape(lat), url.QueryEscape(long), url.QueryEscape(tzname)),
	}, nil
}

func (om *OpenMeteoService) CurrentConditions() (weather.Conditions, error) {
	resp, err := http.Get(om.current_conditions)
	if err != nil {
		return weather.Conditions{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Conditions{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.Conditions{}, err
	}

	var data WeatherResp
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return weather.Conditions{}, err
	}

	return weather.Conditions{
		Temperature: data.Current.Temperature,
		RelHumidity: data.Current.RelativeHumidity,
		WindSpeed:   data.Current.WindSpeed,
		CloudCover:  data.Current.CloudCover,
		WeatherCode: data.Current.WeatherCode,
	}, nil
}

func (om *OpenMeteoService) TemperatureForcast24h() (weather.Forcast, error) {
	resp, err := http.Get(om.forcast)
	if err != nil {
		return weather.Forcast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Forcast{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.Forcast{}, err
	}

	var data TemperatureForcastResp
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return weather.Forcast{}, err
	}

	location, err := time.LoadLocation(data.Timezone)
	if err != nil {
		return weather.Forcast{}, err
	}

	layout := "2006-01-02T15:04"
	switch data.HourlyUnits.Time {
	case "iso8601":
		layout = "2006-01-02T15:04"
	default:
		log.Println("unexpected time layout from OpenMeteoService:", data.HourlyUnits.Time)
	}

	// only keep future data points (ie. time after now)
	now := time.Now()

	for i, time_str := range data.Hourly.Time {
		t, err := time.ParseInLocation(layout, time_str, location)
		if err != nil {
			log.Println("failed to parse time string from OpenMeteoService:", time_str)
		}
		if t.After(now) {
			data.Hourly.Temperatures = data.Hourly.Temperatures[i:]
			data.Hourly.Time = data.Hourly.Time[i:]
			break
		}
	}

	// only forcast the next 24hours
	limit := now.Add(24 * time.Hour)

	for i, time_str := range data.Hourly.Time {
		t, err := time.ParseInLocation(layout, time_str, location)
		if err != nil {
			log.Println("failed to parse time string from OpenMeteoService:", time_str)
		}
		if t.After(limit) {
			data.Hourly.Temperatures = data.Hourly.Temperatures[:i]
			data.Hourly.Time = data.Hourly.Time[:i]
			break
		}
	}

	return weather.Forcast{
		Temperatures: data.Hourly.Temperatures,
	}, nil

}
