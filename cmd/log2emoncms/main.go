package main

import (
	"burlo/pkg/dx2w"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	// TODO: create config for EmonCMS settings
	// The following api key is safe to post publicly
	cache_server := flag.String("s", "192.168.50.193:4006", "DX2W modbus state cache server http address:port")
	emoncmsAddr := flag.String("emoncms", "192.168.50.2:8081", "EmonCMS address:port")
	apikey := flag.String("key", "06b4d1fe9f20d74bcdd44cadc0c02fe7", "EmonCMS api key")
	flag.Parse()

	client := NewClient(fmt.Sprintf("http://%s", *emoncmsAddr), *apikey)
	defer client.Close()

	request_url := fmt.Sprintf("http://%s/dx2w/registers", *cache_server)

	for {
		start := time.Now()

		resp, err := http.Get(request_url)
		if err != nil {
			fmt.Println("Error making the request:", err)
			time.Sleep(time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Println("Error reading the response body:", err)
			time.Sleep(time.Second)
			continue
		}

		var results map[string]dx2w.Value
		err = json.Unmarshal(body, &results)
		if err != nil {
			fmt.Printf("Error decoding the JSON response: %v\n", err)
			os.Exit(1)
		}

		log.Println("writing measurements to emoncms")
		client.Post(results)

		// the dx2w/registers cache is updated at most once
		// every 15 seconds
		wait := 15*time.Second - time.Since(start)
		time.Sleep(wait)
	}
}
