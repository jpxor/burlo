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

modbus:
	go build ../services/modbus
	systemctl restart hvac.weather.service

all: actuator controller vthermostat weather modbus

clean:
	rm ./controller
	rm ./vthermostat
	rm ./weather
	rm ./modbus