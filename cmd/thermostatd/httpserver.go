package main

import (
	"burlo/config"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func http_server(ctx context.Context, cfg config.ServiceConf) {
	_, port := splitAddr(cfg.ServiceHTTPAddresses.Thermostat)

	mux := http.NewServeMux()
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Handler:      mux,
	}
	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()

	mux.HandleFunc("PUT /thermostat/{id}/name", PutThermostatName)
	mux.HandleFunc("PUT /thermostat/{id}/setpoint", PutThermostatSetpoint)
	mux.HandleFunc("GET /thermostats", GetThermostats)

	fmt.Println("http server listening on", server.Addr)
	server.ListenAndServe()
}

func GetThermostats(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	bytes, err := json.MarshalIndent(thermostats, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
	w.WriteHeader(http.StatusOK)
}

func PutThermostatName(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	mutex.Lock()
	defer mutex.Unlock()

	id := r.PathValue("id")
	tstat, ok := thermostats[id]
	if !ok {
		http.Error(w, "unknown id", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tstat.Name = req.Name
	thermostats[id] = tstat

	publish(tstat)
	w.WriteHeader(http.StatusOK)
}

func PutThermostatSetpoint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HeatSetpoint float32 `json:"heat_setpoint,omitempty"`
		CoolSetpoint float32 `json:"cool_setpoint,omitempty"`
	}
	mutex.Lock()
	defer mutex.Unlock()

	id := r.PathValue("id")
	tstat, ok := thermostats[id]
	if !ok {
		http.Error(w, "unknown id", http.StatusBadRequest)
		return
	}
	// set current values in case request ommits one
	req.HeatSetpoint = tstat.HeatSetpoint
	req.CoolSetpoint = tstat.CoolSetpoint

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tstat.HeatSetpoint = req.HeatSetpoint
	tstat.CoolSetpoint = req.CoolSetpoint

	thermostats[id] = tstat
	publish(tstat)

	w.WriteHeader(http.StatusOK)
}

func splitAddr(addr string) (string, string) {
	host, port, _ := strings.Cut(addr, ":")
	return host, port
}
