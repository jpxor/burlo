package main

import (
	"burlo/config"
	"burlo/pkg/models/controller"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func httpserver(ctx context.Context, cfg config.ServiceConf) {
	_, port := splitAddr(cfg.ServiceHTTPAddresses.Dashboard)

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

	mux.HandleFunc("GET /ws", AcceptWebsocket())
	mux.HandleFunc("GET /dashboard", RenderDashboard())
	mux.HandleFunc("GET /{file}", ServeFile())
	mux.HandleFunc("/", RedirectTo("/dashboard"))

	for {
		fmt.Println("http server listening on", server.Addr)
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			break
		}
		fmt.Println(err)
	}
}

func RenderDashboard() http.HandlerFunc {
	var wwwpath = "./www"
	_, err := os.Stat(wwwpath)
	if err != nil {
		wwwpath = "./cmd/dashboard/www"
	}
	tmpl, err := template.ParseFiles(
		filepath.Join(wwwpath, "templates/dashboard/main.html"),
		filepath.Join(wwwpath, "templates/dashboard/setpoint.html"),
		filepath.Join(wwwpath, "templates/dashboard/roomstats.html"))
	if err != nil {
		panic(err)
	}
	type PageData struct {
		Title       string
		Heading     string
		Setpoint    SetpointData
		Thermostats []controller.Thermostat
		Unit        string
	}
	data := PageData{
		Title:   "Dashboard",
		Heading: "Dashboard",
		Unit:    "°C",
		Setpoint: SetpointData{
			HeatingSetpoint: 20,
			CoolingSetpoint: 24,
			Mode:            "Heat",
		},
		Thermostats: []controller.Thermostat{
			{Name: "Living Room", Temperature: 20, Humidity: 40},
			{Name: "Office", Temperature: 20, Humidity: 40},
			{Name: "Yoga Room", Temperature: 20, Humidity: 40},
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// reload the templates on each request ONLY in dev
		tmpl, err = template.ParseFiles(
			filepath.Join(wwwpath, "templates/dashboard/main.html"),
			filepath.Join(wwwpath, "templates/dashboard/setpoint.html"),
			filepath.Join(wwwpath, "templates/dashboard/roomstats.html"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// update the data
		err = tmpl.Execute(w, data)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func ServeFile() http.HandlerFunc {
	var path = "./www"
	_, err := os.Stat(path)
	if err != nil {
		path = "./cmd/dashboard/www"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		file := r.PathValue("file")
		urlpath, err := url.JoinPath(path, file)
		if err != nil {
			http.Error(w, "unexpected error", http.StatusInternalServerError)
		}
		http.ServeFile(w, r, urlpath)
	}
}

func RedirectTo(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func splitAddr(addr string) (string, string) {
	host, port, _ := strings.Cut(addr, ":")
	return host, port
}
