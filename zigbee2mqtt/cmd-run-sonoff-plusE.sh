#!/usr/bin/sh
echo "configuring for Sonoff dongle plus E"
ZIGBEE2MQTT_CONFIG_SERIAL_PORT=/dev/ttyUSB0 ZIGBEE2MQTT_CONFIG_ADVANCED_PAN_ID=0xBEEF DEVICE_ID=usb-Silicon_Labs_Sonoff_Zigbee_3.0_USB_Dongle_Plus_0001-if00-port0 docker compose up -d
