package main

import (
	"burlo/config"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

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

	mux.HandleFunc("GET /ws", AcceptWebsocket())
	mux.HandleFunc("GET /dashboard", ServeDashboard())
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

func ServeDashboard() http.HandlerFunc {
	var path = "./www/index.html"
	_, err := os.Stat(path)
	if err != nil {
		path = "./cmd/dashboard/www/index.html"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
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
