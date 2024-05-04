package main

import (
	"context"
	_ "embed"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//go:embed dx2w-modbus.toml
var modbusConf []byte

type global_vars struct {
	waitgroup sync.WaitGroup
}

var global = global_vars{}

func main() {

	printRegs := flag.Bool("print", false, "[debug] just print all register values")
	flag.Parse()

	cfg := ParseConfig(modbusConf)

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
