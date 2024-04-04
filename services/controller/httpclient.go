package main

import (
	protocol "burlo/services/protocols"
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

var urlPhi string
var urlMB string

func initHttpClient(addrPhidgets, addrModbus string) {
	urlPhi = fmt.Sprintf("http://%s%s", addrPhidgets, "/phidgets/digital_out")
	urlMB = fmt.Sprintf("http://%s%s", addrModbus, "/modbus/register/float")
}

func set_digital_out(data protocol.PhidgetDO) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println("[controller] set_digital_out: failed to encode protocol.PhidgetDO data:", err)
		return
	}
	req, err := http.NewRequest("POST", urlPhi, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[controller] set_digital_out: failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Println("[controller] set_digital_out: failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[controller] set_digital_out: failed to read response body:", err)
		}
		log.Println("[controller] set_digital_out: bad status:", resp.StatusCode, string(bodyBytes))
	}
}

func set_modbus_register(data protocol.ModbusReg) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println("[controller] set_modbus_register: failed to encode protocol.PhidgetDO data:", err)
		return
	}
	req, err := http.NewRequest("POST", urlMB, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[controller] set_modbus_register: failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Println("[controller] set_modbus_register: failed to send request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[controller] set_modbus_register: failed to read response body:", err)
		}
		log.Println("[controller] set_modbus_register: bad status:", resp.StatusCode, string(bodyBytes))
	}
}
