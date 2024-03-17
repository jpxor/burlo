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
	"sync"
	"syscall"
	"time"
)

type global_vars struct {
	waitgroup   sync.WaitGroup
	mutex       sync.Mutex
	state       SystemStateV2
	conditions  ControlConditions
	thermostats map[string]services.Thermostat
}

var global = global_vars{
	state: SystemStateV2{
		Circulator{
			initValue(false),
		},
		Heatpump{
			Mode:          initValue("off"),
			TsTemperature: initValue(float32(20)),
			TsCorrection:  initValue(float32(0)),
		},
	},
	conditions: ControlConditions{
		IndoorConditions{},
		OutdoorConditions{},
	},
	thermostats: make(map[string]services.Thermostat),
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

func update_outdoor_conditions(odc OutdoorConditions) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	global.conditions.OutdoorConditions = odc
	global.conditions.OutdoorConditions.LastUpdate = time.Now()
	global.state = system_update(global.state, global.conditions)
}

func update_indoor_conditions(tstat services.Thermostat) {
	global.mutex.Lock()
	defer global.mutex.Unlock()

	// update thermostat cache, recalculate
	// average setpoint error
	global.thermostats[tstat.ID] = tstat

	var idc IndoorConditions
	for _, tstat := range global.thermostats {
		idc.IndoorAirTempMax = max(idc.IndoorAirTempMax, tstat.State.Temperature)
		idc.DewPoint = max(idc.DewPoint, tstat.State.DewPoint)
		switch global.state.Heatpump.Mode.Value {
		case "heat":
			idc.SetpointError += tstat.State.Temperature - tstat.HeatSetpoint
		case "cool":
			idc.SetpointError += tstat.State.Temperature - tstat.CoolSetpoint
		}
	}
	idc.SetpointError /= float32(len(global.thermostats))

	global.conditions.IndoorConditions = idc
	global.conditions.IndoorConditions.LastUpdate = time.Now()
	global.state = system_update(global.state, global.conditions)
}
