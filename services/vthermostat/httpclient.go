package main

import (
	"burlo/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var controller_upate_url string
var ctr_update_path = "/controller/thermostat/update"

var httpclient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	},
}

func load_controller_addr(cfg config.ServiceConf) {
	controller_upate_url = fmt.Sprintf("http://%s%s", cfg.ServiceHTTPAddresses.Controller, ctr_update_path)
}

func notify_controller(tstat Thermostat) {
	payload, err := json.Marshal(tstat)
	if err != nil {
		log.Println("[notify_controller] failed to encode thermostat data:", err)
		return
	}
	req, err := http.NewRequest("POST", controller_upate_url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[notify_controller] failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Println("[notify_controller] failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[notify_controller] failed to read response body:", err)
		}
		log.Println("[notify_controller] bad status:", resp.StatusCode, string(bodyBytes))
		return
	}
}
