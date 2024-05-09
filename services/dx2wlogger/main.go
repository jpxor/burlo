package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"burlo/config"
	"burlo/pkg/dx2w"
)

var global_mutex sync.Mutex
var register_map = make(map[string]dx2w.Value)

func main() {

	config_path := flag.String("c", "./services.toml", "Path to the services config file")
	flag.Parse()

	cfg := config.LoadV2(*config_path)

	port := config.GetPort(cfg.ServiceHTTPAddresses.Dx2Wlogger)
	go http_server(port)

	// DX2W Modbus TCP device
	dev := dx2w.TCPDevice{
		Url: fmt.Sprintf("tcp://%s", cfg.Dx2WModbus.TCPAddress),
		Id:  cfg.Dx2WModbus.DeviceID,
	}

	var timer_60min time.Time
	var timer_10min time.Time
	var timer_1min time.Time

	for {

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

		// wait the shortest interval
		wait := (15 * time.Second) - time.Since(start)
		time.Sleep(wait)
	}
}

func update_register_map(results map[string]dx2w.Value) {
	global_mutex.Lock()
	defer global_mutex.Unlock()
	for k, v := range results {
		register_map[k] = v
	}
}
