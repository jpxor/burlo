package main

import (
	"burlo/pkg/lockbox"
	. "burlo/services/controller/model"
	services "burlo/services/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type global_vars struct {
	state     *lockbox.LockBox[SystemState]
	waitgroup sync.WaitGroup
}

var global = global_vars{
	state: lockbox.New(SystemState{
		Thermostats: make(map[string]services.Thermostat),
	}),
}

func main() {
	log.Println("Running controller service")

	global.waitgroup.Add(1)
	go controller_http_server()

	global.waitgroup.Wait()
}

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
		state, lbk := global.state.Take()
		defer global.state.Release(lbk)
		w.Write([]byte(fmt.Sprintf("%+v\r\n", state.ControlState)))
		w.Write([]byte(fmt.Sprintf("%+v\r\n", state.ControlConditions)))
		w.Write([]byte(fmt.Sprintf("%+v\r\n", state.Thermostats)))
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

func update_outdoor_conditions(conditions OutdoorConditions) {
	log.Println(conditions)
}

func update_indoor_conditions(tstat services.Thermostat) {
	state, lbk := global.state.Take()

	state.Thermostats[tstat.ID] = tstat

	var idc IndoorConditions
	for _, tstat := range state.Thermostats {
		idc.IndoorAirTempMax = max(idc.IndoorAirTempMax, tstat.State.Temperature)
		idc.DewPoint = max(idc.DewPoint, tstat.State.DewPoint)
		switch state.Mode {
		case Heat:
			idc.SetpointError += tstat.State.Temperature - tstat.HeatSetpoint
		case Cool:
			idc.SetpointError += tstat.State.Temperature - tstat.CoolSetpoint
		}
	}

	// average setpoint error
	idc.SetpointError /= float32(len(state.Thermostats))
	state.IndoorConditions = idc

	state.ControlState = UpdateControls(state.ControlState, ControlConditions{
		state.IndoorConditions,
		state.OutdoorConditions,
	})
	apply(state.ControlState)
	global.state.Put(state, lbk)
}

func apply(output ControlState) {
	log.Println("apply:", output)
}
