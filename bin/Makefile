#########

controller:
	go build ../services/controller

vthermostat:
	go build ../services/vthermostat

weather:
	go build ../services/weather

all: controller vthermostat weather
	
#########

deploy_actuator:
	systemctl restart hvac.actuator.phidgets.service

deploy_controller: controller
	systemctl restart hvac.controller.service

deploy_vthermostat: vthermostat
	systemctl restart hvac.vthermostat.service

deploy_weather: weather
	systemctl restart hvac.weather.service

deploy_all: deploy_actuator deploy_controller deploy_weather deploy_vthermostat

##########