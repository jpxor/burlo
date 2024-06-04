#!/usr/bin/sh
cd /usr/userapps/hvac-controller/burlo
go build ./cmd/controllerd
./controllerd -c ./config/services.toml
