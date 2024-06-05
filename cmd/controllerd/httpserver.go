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

func splitAddr(addr string) (string, string) {
	host, port, _ := strings.Cut(addr, ":")
	return host, port
}

func httpserver(ctx context.Context, cfg config.ServiceConf) {
	_, port := splitAddr(cfg.ServiceHTTPAddresses.Controller)

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

	mux.HandleFunc("GET /controller/state", GetControllerState())
	mux.HandleFunc("GET /controller/emoncms", GetEmoncmsInputs())
	for {
		fmt.Println("http server listening on", server.Addr)
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			break
		}
		fmt.Println(err)
	}
}

func GetControllerState() http.HandlerFunc {
	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}
	return func(w http.ResponseWriter, r *http.Request) {
		inputMutex.Lock()
		defer inputMutex.Unlock()
		w.Write(jsonBytes(State{
			inputs,
			currentState,
		}))
	}
}

func GetEmoncmsInputs() http.HandlerFunc {
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
		inputMutex.Lock()
		defer inputMutex.Unlock()
		var state = make(map[string]float32)
		state["tstat_call"] = bool2float(currentState.ZoneCall)
		state["hp_cooling_mode"] = bool2float(currentState.DX2W.Mode == DX2W_COOL)
		state["outdoor_air_temp"] = inputs.Outdoor.Temperature
		state["indoor_dewpoint"] = inputs.Indoor.Dewpoint
		state["indoor_air_temp"] = inputs.Indoor.Temperature
		w.Write(jsonBytes(state))
	}
}
