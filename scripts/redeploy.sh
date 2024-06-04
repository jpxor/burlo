#!/usr/bin/sh
systemctl restart hvac.actuator.phidgets.service
systemctl restart hvac.controller.service
systemctl restart hvac.vthermostat.service
systemctl restart hvac.weather.service
systemctl restart hvac.modbus.service
