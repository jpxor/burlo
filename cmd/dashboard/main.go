package main

import (
	"burlo/config"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	cfg := config.LoadV2(*configPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("started")
	defer fmt.Println("stopped")

	dashboard := NewDashboard()
	go dashboard.mqttListener(ctx, cfg)
	go dashboard.httpserver(ctx, cfg)

	// waits for signal
	<-ctx.Done()
}
