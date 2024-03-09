package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	tmpl_name_change := template.Must(template.ParseFiles("./www/templates/name-change-form.html"))
	tmpl_name_change_confirm := template.Must(template.ParseFiles("./www/templates/name-change-confirm.html"))

	mux.HandleFunc("GET /thermostats/state", GetThermostatsState(tmpl_state))
	mux.HandleFunc("GET /thermostats/history", GetThermostatsHistory(tmpl_history))
	mux.HandleFunc("GET /", GetIndex)

	mux.HandleFunc("GET /thermostat/{id}/name-change-form", GetThermostatNameChangeForm(tmpl_name_change))
	mux.HandleFunc("PUT /thermostat/{id}/name", PutThermostatName(tmpl_name_change_confirm))
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

func GetThermostatNameChangeForm(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEVEL {
			// reload the template on each request when in development
			tmpl = template.Must(template.ParseFiles("./www/templates/name-change-form.html"))
		}
		id := r.PathValue("id")
		thermostats, lbk := global.thermostats.Take()
		defer global.thermostats.Release(lbk)

		tstat, ok := thermostats[id]
		if !ok {
			http.Error(w, "unknown thermostat id", http.StatusBadRequest)
			return
		}
		tmpl.Execute(w, tstat)
	}
}

func PutThermostatName(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEVEL {
			// reload the template on each request when in development
			tmpl = template.Must(template.ParseFiles("./www/templates/name-change-confirm.html"))
		}
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}
		id := r.PathValue("id")
		name := r.Form.Get("name")

		if len(name) == 0 {
			http.Error(w, "missing request param 'name'", http.StatusBadRequest)
			return
		}
		thermostats, lbk := global.thermostats.Take()
		defer global.thermostats.Release(lbk)

		tstat, ok := thermostats[id]
		if !ok {
			http.Error(w, "unknown thermostat id", http.StatusBadRequest)
			return
		}
		tstat.Name = name
		thermostats[id] = tstat
		tmpl.Execute(w, tstat)
	}
}

func PutThermostatSetpoint(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	setpoint_str := r.Form.Get("setpoint")
	if len(setpoint_str) == 0 {
		http.Error(w, "missing request param 'setpoint'", http.StatusBadRequest)
		return
	}

	setpoint, err := strconv.ParseFloat(setpoint_str, 32)
	if err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	thermostats, lbk := global.thermostats.Take()
	defer global.thermostats.Release(lbk)

	tstat, ok := thermostats[id]
	if !ok {
		http.Error(w, "unknown thermostat id", http.StatusBadRequest)
		return
	}
	if setpoint < 0 || setpoint > 30 {
		http.Error(w, "invalid setpoint, must be within range [0-30] degrees Celcius", http.StatusBadRequest)
		return
	}
	tstat.Setpoint = float32(setpoint)
	thermostats[id] = tstat
	w.WriteHeader(http.StatusOK)
}
