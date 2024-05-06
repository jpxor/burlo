actuator:
	systemctl restart hvac.actuator.phidgets.service

controller:
	go build ../services/controller
	systemctl restart hvac.controller.service

vthermostat:
	go build ../services/vthermostat
	systemctl restart hvac.vthermostat.service

weather:
	go build ../services/weather
	systemctl restart hvac.weather.service

dx2wlogger:
	go build ../services/dx2wlogger
	systemctl restart hvac.modbus.service

all: actuator controller vthermostat weather dx2wlogger

clean:
	rm ./controller
	rm ./vthermostat
	rm ./weather
	rm ./dx2wlogger