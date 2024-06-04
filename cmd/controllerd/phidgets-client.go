package main

import (
	protocol "burlo/services/protocols"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var phidgets_client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	},
}

var uriDO string
var uriVO string

func initPhidgetsClient(addrPhidgets string) {
	uriDO = fmt.Sprintf("http://%s%s", addrPhidgets, "/phidgets/digital_out")
	uriVO = fmt.Sprintf("http://%s%s", addrPhidgets, "/phidgets/voltage_out")
}

func set_digital_out(data protocol.PhidgetDO) {
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("[controller] set_digital_out: failed to encode protocol.PhidgetDO:", err)
		return
	}
	req, err := http.NewRequest("POST", uriDO, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("[controller] set_digital_out: failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := phidgets_client.Do(req)
	if err != nil {
		fmt.Println("[controller] set_digital_out: failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[controller] set_digital_out: failed to read response body:", err)
		}
		fmt.Println("[controller] set_digital_out: bad status:", resp.StatusCode, string(bodyBytes))
	}
}

func set_voltage_out(data protocol.PhidgetVO) {
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("[controller] set_voltage_out: failed to encode protocol.PhidgetVO:", err)
		return
	}
	req, err := http.NewRequest("POST", uriVO, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("[controller] set_voltage_out: failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := phidgets_client.Do(req)
	if err != nil {
		fmt.Println("[controller] set_voltage_out: failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[controller] set_voltage_out: failed to read response body:", err)
		}
		fmt.Println("[controller] set_voltage_out: bad status:", resp.StatusCode, string(bodyBytes))
	}
}
