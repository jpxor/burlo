package main

import (
	"context"
	"flag"
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

	printRegs := flag.Bool("print", false, "[debug] just print all register values")
	confpath := flag.String("c", "./dx2w-modbus.toml", "path to the modbus register configurations file")
	flag.Parse()

	cfg := LoadConfig(*confpath)

	if *printRegs {
		printRegisters(cfg)
		os.Exit(0)
	}

	// start services
	cache := NewAutoCache(cfg, 20*time.Second)
	defer cache.Stop()

	cache.Subscribe(init_energy_monitor())

	// wait for signal before exiting
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
}
