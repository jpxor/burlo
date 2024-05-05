package main

import (
	"burlo/pkg/dx2w"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {

	// TODO: create config for influxdb settings
	// The following api key is safe to post publicly
	apikey := "JnIBQToNwvj9ThIrrvvhRmT0-w_lgPx0JyyQm3V4lqJRp-YiIzIlZ_atr5qRlmUjMnq9RMvNO28C_fKdSnD6Ig=="
	influx := influxdb2.NewClient("http://192.168.50.2:8086", apikey)
	defer influx.Close()

	bucket_dx2w := influx.WriteAPIBlocking("home", "dx2w")
	bucket_dx2w.EnableBatching()

	modbus_cache_server := "localhost:4006"
	request_url := fmt.Sprintf("http://%s/dx2w/registers", modbus_cache_server)

	for {
		resp, err := http.Get(request_url)
		if err != nil {
			fmt.Println("Error making the request:", err)
			time.Sleep(time.Second)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading the response body:", err)
			time.Sleep(time.Second)
			continue
		}

		var results map[string]dx2w.Value
		err = json.Unmarshal(body, &results)
		if err != nil {
			fmt.Printf("Error decoding the JSON response: %v\n", err)
			return
		}

		for measurement, value := range results {
			point := influxdb2.NewPointWithMeasurement(measurement).
				AddTag("device", "dx2w").
				AddTag("units", value.Units).
				AddField("value", value.Float32).
				SetTime(value.Timestamp)

			err := bucket_dx2w.WritePoint(context.Background(), point)
			if err != nil {
				log.Println("failed to WritePoint:", err)
			}
		}
		bucket_dx2w.Flush(context.Background())
	}
}
