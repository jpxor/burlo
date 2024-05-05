package main

import (
	"context"
	"log"
	"sync"
	"time"

	"burlo/pkg/dx2w"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var global_mutex sync.Mutex
var register_map = make(map[string]dx2w.Value)

func main() {
	go http_server()

	// DX2W Modbus TCP device
	dev := dx2w.TCPDevice{
		Url: "tcp://192.168.50.60:502",
		Id:  200,
	}

	// TODO: create config for influxdb settings
	// The following api key is safe to post publicly
	apikey := "JnIBQToNwvj9ThIrrvvhRmT0-w_lgPx0JyyQm3V4lqJRp-YiIzIlZ_atr5qRlmUjMnq9RMvNO28C_fKdSnD6Ig=="
	influx := influxdb2.NewClient("http://192.168.50.2:8086", apikey)
	defer influx.Close()

	bucket_dx2w := influx.WriteAPIBlocking("home", "dx2w")
	bucket_dx2w.EnableBatching()

	var timer_60min time.Time
	var timer_10min time.Time
	var timer_1min time.Time

	for {

		global_mutex.Lock()
		start := time.Now()

		// build up the fields list based on which fields are
		// to be read this iteration
		var fields []string

		// always add fields from shortest interval
		fields = append(fields, fields_15sec_interval...)

		if time.Since(timer_60min) > time.Hour {
			fields = append(fields, fields_static...)
			timer_60min = time.Now()
		}

		if time.Since(timer_10min) > 10*time.Minute {
			fields = append(fields, fields_slow...)
			timer_10min = time.Now()
		}

		if time.Since(timer_1min) > time.Minute {
			fields = append(fields, fields_1min_interval...)
			timer_1min = time.Now()
		}

		client := dx2w.NewWithFields(dev, fields)
		results := client.ReadAll()

		update_register_map(results)
		global_mutex.Unlock()

		logResults(results, bucket_dx2w)
		bucket_dx2w.Flush(context.Background())

		// wait the shortest interval
		wait := (15 * time.Second) - time.Since(start)
		time.Sleep(wait)
	}
}

func logResults(results map[string]dx2w.Value, influx_bucket api.WriteAPIBlocking) {
	for measurement, value := range results {
		point := influxdb2.NewPointWithMeasurement(measurement).
			AddTag("device", "dx2w").
			AddTag("units", value.Units).
			AddField("value", value.Float32).
			SetTime(value.Timestamp)

		err := influx_bucket.WritePoint(context.Background(), point)
		if err != nil {
			log.Println("failed to WritePoint:", err)
		}
	}
}

func update_register_map(results map[string]dx2w.Value) {
	for k, v := range results {
		register_map[k] = v
	}
}
