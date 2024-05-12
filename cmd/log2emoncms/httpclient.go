package main

import (
	"burlo/pkg/dx2w"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
)

type Client struct {
	address string
	key     string
}

func NewClient(addr, key string) *Client {
	return &Client{
		address: addr,
		key:     key,
	}
}

func (c *Client) Close() {}

func (c *Client) Post(data map[string]dx2w.Value) {

	bytes, err := json.Marshal(filter(format(data)))
	if err != nil {
		log.Println("failed to marshal:", err)
		return
	}

	request := fmt.Sprintf("%s/input/post/dx2w?apikey=%s&fulljson=%s", c.address, c.key, string(bytes))

	resp, err := http.Get(request)
	if err != nil {
		log.Println("failed http request:", err)
		return
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
}

func format(data map[string]dx2w.Value) map[string]float32 {
	ret := make(map[string]float32)
	for k, v := range data {
		ret[k] = v.Float32
	}
	return ret
}

func filter(data map[string]float32) map[string]float32 {
	var selected = []string{
		"COMPRESSOR_CALL",
		"HP_CIRCULATOR",
		"HP_INPUT_KW",
		"HP_OUTPUT_KW",
		"BUFFER_FLOW",
		"BUFFER_TANK_SETPOINT",
		"BUFFER_TANK_TEMPERATURE",
		"HP_ENTERING_WATER_TEMP",
		"HP_EXITING_WATER_TEMP",
		"OUTSIDE_AIR_TEMP",
		"MIX_WATER_TEMP",
		"RETURN_WATER_TEMP",
	}
	ret := make(map[string]float32)
	for k, v := range data {
		if slices.Contains(selected, k) {
			ret[k] = v
		}
	}
	return ret
}
