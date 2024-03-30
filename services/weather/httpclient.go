package main

import (
	"burlo/services/protocols/controller"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var httpclient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	},
}

var controller_server_url string

func initHttpClient(srvaddr string) {
	controller_server_url = fmt.Sprintf("http://%s/controller/weather/update", srvaddr)
}

func notify_controller(conditions controller.OutdoorConditions) {
	payload, err := json.Marshal(conditions)
	if err != nil {
		log.Println("[weather] notify_controller: failed to encode thermostat data:", err)
		return
	}
	req, err := http.NewRequest("POST", controller_server_url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[weather] notify_controller: failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Println("[weather] notify_controller: failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[weather] notify_controller: failed to read response body:", err)
		}
		log.Println("[weather] notify_controller: bad status:", resp.StatusCode, string(bodyBytes))
	}
}
