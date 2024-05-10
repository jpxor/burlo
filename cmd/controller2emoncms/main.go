package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {

	emoncmsAddr := flag.String("emoncms", "192.168.50.2:8081", "EmonCMS address:port")
	emoncmsKey := flag.String("ekey", "06b4d1fe9f20d74bcdd44cadc0c02fe7", "EmonCMS api key")

	controllerAddr := flag.String("controller", "192.168.50.193:4005", "Burlo Controller Service address:port")

	flag.Parse()

	emoncms := NewClient(fmt.Sprintf("http://%s", *emoncmsAddr), *emoncmsKey)
	defer emoncms.Close()

	url := fmt.Sprintf("http://%s/controller/state", *controllerAddr)

	for {
		start := time.Now()

		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("failed get request", err)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("failed to read body", err)
		}
		resp.Body.Close()

		var results map[string]float32
		err = json.Unmarshal(data, &results)
		if err != nil {
			log.Fatal("failed to parse json", err)
		}

		log.Println("writing measurements to emoncms")
		emoncms.Post(results)

		wait := 15*time.Second - time.Since(start)
		time.Sleep(wait)
	}
}
