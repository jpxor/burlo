package weathergcca

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type FeatureCollection struct {
	Type           string    `json:"type"`
	Features       []Feature `json:"features"`
	NumberMatched  int       `json:"numberMatched"`
	NumberReturned int       `json:"numberReturned"`
	Links          []Link    `json:"links"`
	TimeStamp      string    `json:"timeStamp"`
}

type Feature struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Properties struct {
	ID                     string `json:"id"`
	AQHIType               string `json:"aqhi_type"`
	ForecastType           string `json:"forecast_type"`
	LocationNameEN         string `json:"location_name_en"`
	LocationNameFR         string `json:"location_name_fr"`
	LocationID             string `json:"location_id"`
	PublicationDatetime    string `json:"publication_datetime"`
	ForecastDatetimeTextEN string `json:"forecast_datetime_text_en"`
	ForecastDatetimeTextFR string `json:"forecast_datetime_text_fr"`
	ForecastDatetime       string `json:"forecast_datetime"`
	AQHI                   int    `json:"aqhi"`
}

type Link struct {
	Type  string `json:"type"`
	Rel   string `json:"rel"`
	Title string `json:"title"`
	Href  string `json:"href"`
}

type AqhiForecast struct {
	AQHI []int
	Time []time.Time
}

func GetAqhiForecast(nhours int, location string) (AqhiForecast, error) {
	qlimit := 200
	uri := fmt.Sprintf("https://api.weather.gc.ca/collections/aqhi-forecasts-realtime/items?lang=en&limit=%d&location_name_en=%s&f=json", qlimit, location)

	resp, err := http.Get(uri)
	if err != nil {
		return AqhiForecast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AqhiForecast{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return AqhiForecast{}, err
	}

	var data FeatureCollection
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return AqhiForecast{}, err
	}

	// only keep future data points (ie. time after now)
	now := time.Now()

	layout := "2006-01-02T15:04"

	for i, feat := range data.Features {
		t, err := time.Parse(layout, feat.Properties.ForecastDatetime[:16])
		if err != nil {
			log.Println("failed to parse time string:", feat.Properties.ForecastDatetime)
		}
		if t.After(now) {
			data.Features = data.Features[i:]
			break
		}
	}

	// only forecast the next nhours
	limit := now.Add(time.Duration(nhours) * time.Hour)

	for i, feat := range data.Features {
		t, err := time.Parse(layout, feat.Properties.ForecastDatetime[:16])
		if err != nil {
			log.Println("failed to parse time string:", feat.Properties.ForecastDatetime)
		}
		if t.After(limit) {
			data.Features = data.Features[:i]
			break
		}
	}

	forecast := AqhiForecast{}
	for _, feat := range data.Features {
		t, _ := time.Parse(layout, feat.Properties.ForecastDatetime[:16])
		forecast.AQHI = append(forecast.AQHI, feat.Properties.AQHI)
		forecast.Time = append(forecast.Time, t)
	}
	return forecast, nil
}
