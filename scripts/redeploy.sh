#!/usr/bin/sh -x

cd /usr/userapps/hvac-controller/burlo
go build ./cmd/weatherd
go build ./cmd/thermostatd
go build ./cmd/controllerd

systemctl restart hvac.actuator.phidgets.service
systemctl restart hvac.controller.service
systemctl restart hvac.vthermostat.service
systemctl restart hvac.weather.service
systemctl restart hvac.modbus.service