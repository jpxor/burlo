package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var DEVEL = true

func go_gadget_web_app() {
	defer global.waitgroup.Done()

	var port = 4004
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

	// features:
	// - show thermostat history (using apexcharts.js)
	//    - filter by name
	//    - select data type
	// - show thermostats current state
	//    - allow adjusting setpoints
	//    - allow setting name

	tmpl_history := template.Must(template.ParseFiles("./www/templates/history.html"))
	tmpl_state := template.Must(template.ParseFiles("./www/templates/state.html"))

	mux.HandleFunc("GET /thermostats/state", GetThermostatsState(tmpl_state))
	mux.HandleFunc("GET /thermostats/history", GetThermostatsHistory(tmpl_history))
	mux.HandleFunc("GET /", GetIndex)

	mux.HandleFunc("PUT /thermostat/{id}/name", PutThermostatName)
	mux.HandleFunc("PUT /thermostat/{id}/setpoint", PutThermostatSetpoint)

	log.Println("[web_app] started", addr)
	err := server.ListenAndServe()

	if err != http.ErrServerClosed {
		log.Println("[web_app]", err)
	}
	log.Println("[web_app] stopped")
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./www/index.html")
}

func GetThermostatsState(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEVEL {
			// reload the template on each request when in development
			tmpl = template.Must(template.ParseFiles("./www/templates/state.html"))
		}
		thermostats, lbk := global.thermostats.Take()
		err := tmpl.Execute(w, thermostats)
		if err != nil {
			log.Println(err)
		}
		global.thermostats.Release(lbk)
	}
}

func GetThermostatsHistory(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEVEL {
			// reload the template on each request when in development
			tmpl = template.Must(template.ParseFiles("./www/templates/history.html"))
		}
		tmpl.Execute(w, nil)
	}
}

func PutThermostatName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "got path\n")
}

func PutThermostatSetpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "got path\n")
}
