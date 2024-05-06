package main

import (
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

func http_server(port string) {

	mux := http.NewServeMux()
	server := http.Server{
		Addr:         fmt.Sprintf(":%v", port),
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

	mux.HandleFunc("GET /dx2w/registers", GetRegisters())
	mux.HandleFunc("GET /dx2w", GetPageUI())
	mux.HandleFunc("/", CatchAll())

	log.Println("http_server started, port:", port)
	defer log.Println("http_server stopped")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println("[dxw2_logger_http_server]", err)
	}
}

func CatchAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dx2w", http.StatusFound)
	}
}

func GetRegisters() http.HandlerFunc {
	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}
	return func(w http.ResponseWriter, r *http.Request) {
		global_mutex.Lock()
		defer global_mutex.Unlock()
		w.Write(jsonBytes(register_map))
	}
}

// TODO: build html page
func GetPageUI() http.HandlerFunc {
	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}
	return func(w http.ResponseWriter, r *http.Request) {
		global_mutex.Lock()
		defer global_mutex.Unlock()
		w.Write(jsonBytes(register_map))
	}
}
