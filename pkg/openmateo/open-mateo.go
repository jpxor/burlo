package openmateo

import (
	"burlo/pkg/models/weather"
	"burlo/pkg/timezone"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ForcastResp struct {
	Timezone    string `json:"timezone"`
	HourlyUnits struct {
		Time                string `json:"time"`
		Temperature         string `json:"temperature_2m"`
		RelHumidity         string `json:"relative_humidity_2m"`
		ProbPrecipitation   string `json:"precipitation_probability"`
		PrecipitationAmount string `json:"precipitation"`
		CloudCover          string `json:"cloud_cover"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                []string  `json:"time"`
		Temperatures        []float32 `json:"temperature_2m"`
		RelHumidity         []float32 `json:"relative_humidity_2m"`
		ProbPrecipitation   []float32 `json:"precipitation_probability"`
		PrecipitationAmount []float32 `json:"precipitation"`
		CloudCover          []float32 `json:"cloud_cover"`
	} `json:"hourly"`
}

type CurrentResp struct {
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
		Precipitation    string `json:"precipitation"`
	} `json:"current_units"`
	Current struct {
		Time             string  `json:"time"`
		Interval         int     `json:"interval"`
		WeatherCode      int32   `json:"weather_code"`
		Temperature      float32 `json:"temperature_2m"`
		RelativeHumidity float32 `json:"relative_humidity_2m"`
		CloudCover       float32 `json:"cloud_cover"`
		WindSpeed        float32 `json:"wind_speed_10m"`
		Precipitation    float32 `json:"precipitation"`
	} `json:"current"`
}

type OpenMeteoService struct {
	forcast string
	current string
}

func New(lat, long string) (*OpenMeteoService, error) {
	tzname, err := timezone.FromCoordinates(lat, long)
	if err != nil {
		return nil, err
	}
	return &OpenMeteoService{
		forcast: buildURL(lat, long, tzname, "hourly=temperature_2m,relative_humidity_2m,precipitation_probability,precipitation,cloud_cover&forecast_days=3"),
		current: buildURL(lat, long, tzname, "current=temperature_2m,relative_humidity_2m,precipitation,weather_code,cloud_cover"),
	}, nil
}

func buildURL(lat, long, tzname, query string) string {
	lat = url.QueryEscape(lat)
	long = url.QueryEscape(long)
	tzname = url.QueryEscape(tzname)
	return fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&timezone=%s&%s", lat, long, tzname, query)
}

func (om *OpenMeteoService) CurrentConditions() (weather.Current, error) {
	resp, err := http.Get(om.current)
	if err != nil {
		return weather.Current{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Current{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.Current{}, err
	}

	var data CurrentResp
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return weather.Current{}, err
	}

	return weather.Current{
		Temperature:   data.Current.Temperature,
		RelHumidity:   data.Current.RelativeHumidity,
		WindSpeed:     data.Current.WindSpeed,
		CloudCover:    data.Current.CloudCover,
		Precipitation: data.Current.Precipitation,
		WeatherCode:   data.Current.WeatherCode,
	}, nil
}

func (om *OpenMeteoService) Forcast24h() (weather.Forcast, error) {
	resp, err := http.Get(om.forcast)
	if err != nil {
		return weather.Forcast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return weather.Forcast{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather.Forcast{}, err
	}

	var data ForcastResp
	err = json.Unmarshal(bytes, &data)
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
		log.Println("unexpected time layout:", data.HourlyUnits.Time)
	}

	// only keep future data points (ie. time after now)
	now := time.Now()

	for i, time_str := range data.Hourly.Time {
		t, err := time.ParseInLocation(layout, time_str, location)
		if err != nil {
			log.Println("failed to parse time string from OpenMeteoService:", time_str)
		}
		if t.After(now) {
			data.Hourly.Time = data.Hourly.Time[i:]
			data.Hourly.CloudCover = data.Hourly.CloudCover[i:]
			data.Hourly.RelHumidity = data.Hourly.RelHumidity[i:]
			data.Hourly.Temperatures = data.Hourly.Temperatures[i:]
			data.Hourly.ProbPrecipitation = data.Hourly.ProbPrecipitation[i:]
			data.Hourly.PrecipitationAmount = data.Hourly.PrecipitationAmount[i:]
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
			data.Hourly.Time = data.Hourly.Time[:i]
			data.Hourly.CloudCover = data.Hourly.CloudCover[:i]
			data.Hourly.RelHumidity = data.Hourly.RelHumidity[:i]
			data.Hourly.Temperatures = data.Hourly.Temperatures[:i]
			data.Hourly.ProbPrecipitation = data.Hourly.ProbPrecipitation[:i]
			data.Hourly.PrecipitationAmount = data.Hourly.PrecipitationAmount[:i]
			break
		}
	}

	return weather.Forcast{
		Temperature:         data.Hourly.Temperatures,
		RelHumidity:         data.Hourly.RelHumidity,
		ProbPrecipitation:   data.Hourly.ProbPrecipitation,
		PrecipitationAmount: data.Hourly.PrecipitationAmount,
		CloudCover:          data.Hourly.CloudCover,
	}, nil

}
