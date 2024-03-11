package main

import (
	"burlo/model"
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
	waitgroup sync.WaitGroup
}

var global = global_vars{}

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
	mux.HandleFunc("POST /controller/update", PostControllerUpdate())
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
		w.Write([]byte("ACK"))
	}
}

type ControllerUpdate struct {
}

func PostControllerUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data model.Thermostat
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		controller_update(data)
		w.Write([]byte(fmt.Sprintf("%+v\r\n", data)))
	}
}

func controller_update(updated model.Thermostat) {
	log.Println(updated)
}
