package main

import (
	. "burlo/services/controller/model"
	services "burlo/services/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func controller_http_server() {
	defer global.waitgroup.Done()

	var port = 4005
	var addr = fmt.Sprintf(":%d", port)

	mux := http.NewServeMux()
	server := http.Server{
		Addr:         addr,
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

	mux.HandleFunc("GET /controller/state", GetControllerState())
	mux.HandleFunc("POST /controller/thermostat/update", PostThermostatUpdate())
	mux.HandleFunc("POST /controller/weather/update", PostWeatherUpdate())
	mux.HandleFunc("/", CatchAll())

	log.Println("[ctrl_server] started", addr)
	err := server.ListenAndServe()

	if err != http.ErrServerClosed {
		log.Println("[ctrl_server]", err)
	}
	log.Println("[ctrl_server] stopped")
}

func CatchAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Proto, r.Method, r.RequestURI)
		w.WriteHeader(http.StatusNotFound)
	}
}

func GetControllerState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		global.mutex.Lock()
		defer global.mutex.Unlock()
		w.Write([]byte(fmt.Sprintf("%+v\r\n", global.state)))
		w.Write([]byte(fmt.Sprintf("%+v\r\n", global.conditions)))
		w.Write([]byte(fmt.Sprintf("%+v\r\n", global.thermostats)))
	}
}

func PostThermostatUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tstat services.Thermostat
		err := json.NewDecoder(r.Body).Decode(&tstat)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		update_indoor_conditions(tstat)
		w.Write([]byte("ACK"))
	}
}

func PostWeatherUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var conditions OutdoorConditions
		err := json.NewDecoder(r.Body).Decode(&conditions)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		update_outdoor_conditions(conditions)
		w.Write([]byte("ACK"))
	}
}
