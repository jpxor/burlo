# burlo
IOT/Controls for hydronic heating & cooling with air-to-water heatpump

![alt text](burlo.png)

## Sensors

zigbee sensors in each room
 - temperature, humidity
 - zigbee2mqtt writes data to mqtt
 - mqtt message broker (mosquito) pushes data to virtual thermostat service

outdoor weather service
 - periodically polls external weather api (open-mateo)
 - outdoor temperature (current and low/average/high over next 24h) is pushed to controller

## Virtual Thermostats

 - configure heating and cooling setpoints
 - view thermostat data via web interface
 - receives data from zigbee sensors, records history, pushes data to controller

## Controller

 - receives a summary of current indoor & outdoor conditions
 - controls heatpump mode: heating/cooling
 - controls zone pumps: on/off
 - sends dewpoint temperature to heatpump during cooling
 - keeps history of system state, view via web interface


## Actuators

Phidgets relay is used to signal the zone pump via Taco Zone Controller the same way a thermostat would.
https://www.phidgets.com/?prodid=720&pcid=39


## Logs
Navigate to the webpage: http://phidgetsbc.local/cgi-bin/system-logs.sh
Access(admin:root)
