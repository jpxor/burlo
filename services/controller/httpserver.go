package main

import (
	protocol "burlo/services/protocols"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func controller_http_server(addr string) {
	defer global.waitgroup.Done()

	mux := http.NewServeMux()
	server := http.Server{
		Addr:         ":4005",
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Handler:      mux,
	}

	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done() // Waits for signal
		server.Shutdown(context.Background())
	}()

	mux.HandleFunc("GET /controller/state", GetControllerStateJSON())
	mux.HandleFunc("POST /controller/thermostat/update", PostThermostatUpdate())
	mux.HandleFunc("POST /controller/weather/update", PostWeatherUpdate())
	mux.HandleFunc("/", GetControllerStateUI())

	log.Println("[controller_http_server] started", addr)
	defer log.Println("[controller_http_server] stopped")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println("[controller_http_server]", err)
	}
}

func CatchAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Proto, r.Method, r.RequestURI)
		w.WriteHeader(http.StatusNotFound)
	}
}

func GetControllerStateUI() http.HandlerFunc {
	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}
	return func(w http.ResponseWriter, r *http.Request) {
		global.mutex.Lock()
		defer global.mutex.Unlock()
		w.Write(jsonBytes(global.Controls))
		w.Write(jsonBytes(global.OutdoorConditions))
		w.Write(jsonBytes(global.IndoorConditions))
		w.Write(jsonBytes(global.thermostats))
	}
}

func GetControllerStateJSON() http.HandlerFunc {
	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}
	bool2float := func(b bool) float32 {
		if b {
			return 1.0
		}
		return 0.0
	}
	return func(w http.ResponseWriter, r *http.Request) {
		global.mutex.Lock()
		defer global.mutex.Unlock()
		var state = make(map[string]float32)
		state["tstat_call"] = bool2float(global.Controls.Circulator.Mode == "ON")
		state["hp_cooling_mode"] = bool2float(global.Controls.Heatpump.Mode == "Cool")
		state["outdoor_air_temp"] = global.OutdoorConditions.OutdoorAirTemp
		state["indoor_dewpoint"] = global.IndoorConditions.DewPoint
		state["indoor_air_temp"] = global.IndoorConditions.IndoorAirTempMax
		w.Write(jsonBytes(state))
	}
}

func PostThermostatUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data protocol.Thermostat
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var tstat Thermostat
		tstat.From(data)

		update_indoor_conditions(tstat)
		w.Write([]byte("ACK"))
	}
}

func PostWeatherUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data protocol.OutdoorConditions
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var conditions OutdoorConditions
		conditions.From(data)

		update_outdoor_conditions(conditions)
		w.Write([]byte("ACK"))
	}
}
