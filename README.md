# Burlo: Hydronics Control System and Smart Thermostat

Zone controller and wireless (virtual) thermostats which control hydronic radiant heating/cooling using Zigbee sensors and Phidgets actuators.  Written in Golang and Python using microservice architecture with both HTTP and MQTT messaging.

![system diagram showing software and device component relations](burlo.png)

## Sensors

- zigbee temperature and humidity sensors with a display are used in place of thermostats. They wirelessly transmit the room conditions data to the mqtt server (via zigbee2mqtt service) for processing,
- zigbee temperature and humidity sensors WITHOUT a display are used as humidistats and are placed next to hydronic emmiters.  These are used to measure the dewpoint and ensure no condensation while radiant cooling,
- any sensor type can be used so long as it can write its data to mqtt.

## Virtual Thermostats

- this service subscribes to all the sensor data,
- calculates dewpoint from temp/humidity measurements and calculates setpoint errors, then writes the data back to mqtt in a format the controller understands.
- simple httpserver to allow setting thermostat names, heat setpoint, and cool setpoint,
- httpserver also allows querying current thermostat states.

## Weather service

- periodically polls an availalable weather api to determine current conditions and get a 24hr hourly forecast,
- open-mateo provides current conditions, and forecasted temperatures,
- weather.gc.ca provides current and forecast AQHI (air quality health index),
- the weather service writes the data to mqtt for the controller.

## Controller

- subscribes to thermostat, humidistat, and weather data streams via mqtt,
- processes data:
     - determines 24h high, 24h mean, and 24h low temperatures,
     - determines highest indoor dewpoint,
- uses the data to determine:
     - heatpump mode (HEAT/COOL/OFF),
     - zone controller state (ON/OFF),
     - minimum flow temperature (highest dewpoint),
     - if conditions are right for natural ventilation (open windows),
- posts to the Phidgets service to apply mode (HEAT/COOL), zone state (ON/OFF), and dewpoint (converted to 0-10Vdc signal)
- posts to modbus service to apply heatpump state (ON/OFF),
- posts to NTFY service to send notifications (mode and state changes, suggest windows open/close),
- simple httpserver to allow querying current state (inputs and outputs).

## Phidgets service

- phidgets are physical devices used to programatically interact with the real world,
- there is a relay setup to turn ON/OFF up to four 24vdc signals, and a voltage output setup to provide a 0-10vdc signal,
- the phidgets service is a simple httpserver written in python (to use Phidgets python API client) which accepts http post requests to control the devices,
- the httpserver also allows querying current state of the phidget devices.

This system is built on the Phidgets SBC which comes with its own web server. This isn't really used, but can be access here:
    http://phidgetsbc.local/cgi-bin/system-logs.sh
    Access(admin:root)

## Modbus service

- reads and writes from the heatpump modbus device
- reads are scheduled (every 15 seconds for fast changing data like flow and return temperatures, longer periods for slow changing data like buffer setpoint) and cached to limit the number of simultanous modbus requests,
- simple httpserver to allow reading cached values, and writing control/config registers.

## Monitor service (work in progress)

- reads current state from the controller,
- reads current state of the modbus device (via modbus service),
- posts data to EmonCMS to feed its Heatpump Monitor app (which eventially ends up on heatpumpmonitor.org)

## Dashboard service (not implemented)

- reads current state from the controller,
- reads current state of the modbus device (via modbus service),
- webserver to serve dashboard UI,
- allows user input to configure and control the system, posts to controller and thermostat endpoints).

## Logger (not implemented)

- simple httpserver to query logged events
- maintains history of X days

  
