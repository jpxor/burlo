package main

import (
	"log"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type CachedValue struct {
	index     int
	rawval    uint16
	timestamp time.Time
}

type AutoCache struct {
	sync.Mutex
	Config
	Values    map[string]CachedValue
	poll      time.Duration
	signal    chan interface{}
	Listeners []CacheListener
}

type CacheListener func(*AutoCache)

func NewAutoCache(mbConfig Config, poll time.Duration) *AutoCache {
	ac := &AutoCache{
		Config: mbConfig,
		Values: make(map[string]CachedValue),
		signal: make(chan interface{}),
		poll:   poll,
	}
	go ac.start()
	return ac
}

func (ac *AutoCache) Subscribe(fn CacheListener) {
	ac.Listeners = append(ac.Listeners, fn)
}

func (ac *AutoCache) Publish() {
	log.Println("publishing")
	for _, cb := range ac.Listeners {
		go cb(ac)
	}
}

func (ac *AutoCache) Stop() {
	ac.signal <- 0
}

func (ac *AutoCache) start() {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     ac.Config.DeviceURI,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.Fatalln("Error creating Modbus client:", err)
	}

	err = client.Open()
	for err != nil {
		log.Fatalln("Error opening Modbus client:", err)
	}

	defer client.Close()
	client.SetUnitId(ac.Config.DeviceID)

	// init values
	ac.update(client)

	for {
		select {
		case <-ac.signal:
			return
		case <-time.After(ac.poll):
			ac.update(client)
			ac.Publish()
		}
	}
}

func (ac *AutoCache) update(client *modbus.ModbusClient) {
	ac.Lock()
	defer ac.Unlock()
next:
	for i, register := range ac.Config.Register {
		retry := 3
		for {
			rawval, err := client.ReadRegister(register.Address, modbus.HOLDING_REGISTER)
			if err != nil {
				if retry > 0 {
					retry -= 1
					continue
				}
				log.Println("Failed to update modbus cache, register:", register.Address)
				continue next
			}
			ac.Values[register.Name] = CachedValue{
				index:     i,
				rawval:    rawval,
				timestamp: time.Now(),
			}
			break
		}
	}
}

func (ac *AutoCache) AsFloat32(regname string) float32 {
	ac.Lock()
	defer ac.Unlock()

	if cached, ok := ac.Values[regname]; ok {
		register := ac.Register[cached.index]
		switch register.Type {

		case INT16:
			return float32(int16(cached.rawval)) * register.Factor

		case UINT16:
			return float32(cached.rawval) * register.Factor

		case BOOL:
			if cached.rawval != 0 {
				return 1
			}
		}
	}
	return 0 // default value
}

func (ac *AutoCache) AsBool(regname string) bool {
	ac.Lock()
	defer ac.Unlock()

	if cached, ok := ac.Values[regname]; ok {
		if cached.rawval != 0 {
			return true
		}
	}
	return false // default value
}
