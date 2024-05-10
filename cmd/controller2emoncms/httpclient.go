package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (c *Client) Post(data map[string]float32) {

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("failed to marshal:", err)
		return
	}

	request := fmt.Sprintf("%s/input/post/controller?apikey=%s&fulljson=%s", c.address, c.key, string(bytes))

	resp, err := http.Get(request)
	if err != nil {
		log.Println("failed http request:", err)
		return
	}
	defer resp.Body.Close()
	log.Println(resp.Status)
}
