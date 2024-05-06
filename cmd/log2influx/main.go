package main

import (
	"burlo/pkg/dx2w"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {

	cache_server := flag.String("s", "192.168.50.193:4006", "DX2W Modbus Cache servic http address:port")
	influxAddr := flag.String("influxdb", "192.168.50.2:8086", "Influxdb address:port")
	apikey := flag.String("key", "JnIBQToNwvj9ThIrrvvhRmT0-w_lgPx0JyyQm3V4lqJRp-YiIzIlZ_atr5qRlmUjMnq9RMvNO28C_fKdSnD6Ig==", "Influxdb api token")
	flag.Parse()

	// TODO: create config for influxdb settings
	// The following api key is safe to post publicly
	influx := influxdb2.NewClient(fmt.Sprintf("http://%s", *influxAddr), *apikey)
	defer influx.Close()

	bucket_dx2w := influx.WriteAPIBlocking("home", "dx2w")
	bucket_dx2w.EnableBatching()

	request_url := fmt.Sprintf("http://%s/dx2w/registers", *cache_server)

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
