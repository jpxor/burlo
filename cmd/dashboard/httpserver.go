package main

import (
	"burlo/config"
	"burlo/pkg/models/controller"
	"burlo/pkg/mqtt"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type DashboardServer struct {
	dashboard *Dashboard
	mqttc     *mqtt.Client
}

func (d *Dashboard) httpserver(ctx context.Context, cfg config.ServiceConf) {
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

	ds := DashboardServer{
		dashboard: d,
		mqttc: mqtt.NewClient(mqtt.Opts{
			Context:  ctx,
			Address:  cfg.Mqtt.Address,
			User:     cfg.Mqtt.User,
			Pass:     []byte(cfg.Mqtt.Pass),
			ClientID: "dashboard-publisher",
		}),
	}

	mux.HandleFunc("POST /api/v1/setpoint", ds.PostedSetpoint())
	mux.HandleFunc("GET /dashboard", ds.RenderDashboard())
	mux.HandleFunc("GET /ws", AcceptWebsocket())

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

func (s *DashboardServer) PostedSetpoint() http.HandlerFunc {
	var request struct {
		Adjustment float32 `json:"adjustment"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if request.Adjustment > 1 || request.Adjustment < -1 {
			http.Error(w, "invalid adjustment", http.StatusBadRequest)
			return
		}
		s.adjustSetpoint(request.Adjustment)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *DashboardServer) RenderDashboard() http.HandlerFunc {
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
		Thermostats map[string]controller.Thermostat
		Unit        string
		HostAddr    string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			HostAddr:    "192.168.50.6:4001", // TODO get from config
			Title:       "Dashboard",
			Heading:     "Dashboard",
			Unit:        "°C",
			Setpoint:    s.dashboard.Setpoint,
			Thermostats: s.dashboard.Thermostats,
		}
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

func AcceptWebsocket() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		ws.newConnection(conn)
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(messageType, string(p))
		}
	}
}

func splitAddr(addr string) (string, string) {
	host, port, _ := strings.Cut(addr, ":")
	return host, port
}
