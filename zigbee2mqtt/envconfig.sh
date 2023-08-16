#!/usr/bin/sh
if  [ $1 = "sonoffE" ]; then
    echo "settings for Sonoff dongle plus E"
    export ZIGBEE2MQTT_CONFIG_SERIAL_PORT=/dev/ttyUSB0
    export ZIGBEE2MQTT_CONFIG_ADVANCED_PAN_ID=0xBEEF
    export DEVICE_ID=usb-Silicon_Labs_Sonoff_Zigbee_3.0_USB_Dongle_Plus_0001-if00-port0
elif [ $1 = "sonoffP" ]; then
    echo "settings for Sonoff dongle plus P"
    export ZIGBEE2MQTT_CONFIG_SERIAL_PORT=/dev/ttyACM0
    export ZIGBEE2MQTT_CONFIG_SERIAL_ADAPTER=ezsp
    export ZIGBEE2MQTT_CONFIG_ADVANCED_PAN_ID=0xDEAD
    export DEVICE_ID=usb-ITEAD_SONOFF_Zigbee_3.0_USB_Dongle_Plus_V2_20230505182833-if00
else
    echo "ERROR: device not specified:  sonoffE or sonoffP"
    return -1
fi

echo DEVICE_ID ${DEVICE_ID}
echo ZIGBEE2MQTT_CONFIG_SERIAL_PORT ${ZIGBEE2MQTT_CONFIG_SERIAL_PORT}
echo ZIGBEE2MQTT_CONFIG_SERIAL_ADAPTER ${ZIGBEE2MQTT_CONFIG_SERIAL_ADAPTER}
echo ZIGBEE2MQTT_CONFIG_ADVANCED_PAN_ID ${ZIGBEE2MQTT_CONFIG_ADVANCED_PAN_ID}