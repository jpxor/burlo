package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"burlo/pkg/dx2w"
)

var global_mutex sync.Mutex
var register_map = make(map[string]dx2w.Value)

func main() {

	tcpAddr := flag.String("tcp", "192.168.50.60:502", "Modbus TCP address including port number")
	devId := flag.Uint("d", 200, "Target Modbus server device id")
	port := flag.Uint("p", 4006, "HTTP server port")
	flag.Parse()

	go http_server(*port)

	// DX2W Modbus TCP device
	dev := dx2w.TCPDevice{
		Url: fmt.Sprintf("tcp://%s", *tcpAddr),
		Id:  uint8(*devId),
	}

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

		// wait the shortest interval
		wait := (15 * time.Second) - time.Since(start)
		time.Sleep(wait)
	}
}

func update_register_map(results map[string]dx2w.Value) {
	for k, v := range results {
		register_map[k] = v
	}
}
