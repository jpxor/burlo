#!/usr/bin/sh
cd /usr/userapps/hvac-controller/burlo
go build ./cmd/thermostatd
./thermostatd -c ./config/services.toml
