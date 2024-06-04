#!/usr/bin/sh
cd /usr/userapps/hvac-controller/burlo
go build ./cmd/weatherd
./weatherd -c ./config/services.toml
