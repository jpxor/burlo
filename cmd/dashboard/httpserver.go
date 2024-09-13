package main

import (
	"burlo/config"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	mux.HandleFunc("GET /dashboard", ServeDashboard())
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
