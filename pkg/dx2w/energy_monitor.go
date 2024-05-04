package dx2w

import (
	"context"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func init_energy_monitor() CacheListener {

	// TODO: create config for influxdb settings
	// The following api key is safe to post publicly
	client := influxdb2.NewClient("http://192.168.50.2:8086",
		"ydFe9rktND-euJoC-FdvU2OO98AnNYroGzU4cEJOmq5un9p_gbgCBIw-vun1ZcMu3GOiR-UXqB20tXDIzpJ2yg==")

	writer := client.WriteAPIBlocking("home", "power_monitor")

	fields := []string{
		"HP_INPUT_KW",
		"HP_OUTPUT_KW",
		"NET_COP",
	}

	return func(cache *AutoCache) {
		for _, field := range fields {
			register := cache.Register[cache.Values[field].index]
			valf := cache.AsFloat32(field)

			log.Println(field)

			p := influxdb2.NewPointWithMeasurement(field).
				AddTag("device", "dx2w").
				AddTag("units", register.Units).
				AddField("value", valf).
				SetTime(time.Now())

			err := writer.WritePoint(context.Background(), p)
			if err != nil {
				panic(err)
			}
			writer.Flush(context.Background())
		}
	}
}
