package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func notify_controller(tstat Thermostat) {
	// serializes requests from goroutines
	global.notify_queue <- tstat
}

func run_controller_notify_client() {
	defer global.waitgroup.Done()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpclient := &http.Client{
		Timeout: 10 * time.Second,
	}

	host := "192.168.50.193"
	port := "1883"
	path := "/controller/update"
	url := fmt.Sprintf("http://%s:%s%s", host, port, path)

	log.Println("[ctrl] started controller_notify service")
	for {
		select {
		case tstat := <-global.notify_queue:
			notify_controller_synced(url, httpclient, tstat)

		case <-ctx.Done():
			log.Println("[ctrl] signal caught, exiting...")
			return
		}
	}
}

func notify_controller_synced(url string, client *http.Client, tstat Thermostat) {
	payload, err := json.Marshal(tstat)
	if err != nil {
		log.Printf("[ctrl] failed to encode thermostat data: %s\r\n", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("[ctrl] Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ctrl] Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ctrl] Error reading response body:", err)
		}
		log.Println("[ctrl] ", resp.StatusCode, string(bodyBytes))
		return
	}
	log.Println("[ctrl] controller notified")
}
