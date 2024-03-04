
System components - all running on one machine
-----------------------------------------------

===
SENSORS
===

zigbee sensors in each room
 -temperature, humidity

zigbee2mqtt
 - write temperature & humidity data to mqtt

mqtt server
 - message broker

outdoor weather service
 - temperature, humidity, windspeed
 - write to mqtt
 - web interface for status

===
DATA SERVICE
===

tracks current state of all sensors and outdoor conditions
 - listens on mqtt topics
 - state:
    - sensor name, temperature, humidity, dewpoint, last update
    - outdoor temperature, humidity, dewpoint, windspeed
    - average outdoor temperature over 24hr
 - web interface for status

===
CONTROLLER SERVICE
===

determines:
 - heat pump mode [heat, cool, off]
 - target flow temperature
 - zone circulator state [on, off]

provides controls via web interface:
 - thermostat setpoints
 - heat-cool setpoints
 - weather compensation curve (flow temp, outdoor temp) point.

calculates:
 - thermostat dt


===
ACTUATOR SERVICE
===

sends signals to zone controller and dx2w via phigets interface (python or java)
 - web interface for status


===
config
===

a single configuration file for all services

===
resiliance
===

each service must auto restart when failed