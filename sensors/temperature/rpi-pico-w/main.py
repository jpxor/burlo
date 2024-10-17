import time
import dht
import machine
import json
import asyncio
import onewire
import ds18x20
import time

from microwifi import MicroWifi
from micromqttconf import MicroMqttConf
from machine import Pin

# internal LED
led = Pin("LED", Pin.OUT)

# Initialize DHT11 sensor
sensor = dht.DHT11(machine.Pin(28))

# Initialize DS18B20 sensor
ds_pin = machine.Pin(22)
ds_sensor = ds18x20.DS18X20(onewire.OneWire(ds_pin))
ds_roms = ds_sensor.scan()

# Initialize internal temp sensor
internal_temp_sensor = machine.ADC(4)
conversion_factor = 3.3 / (65535)

# Publish data to MQTT server
def publish_data(client, topic, temperature, humidity):
    payload = f"{{\"temperature\": {temperature}, \"humidity\": {humidity}, \"timestamp\": {time.time()}}}"
    client.publish(topic, payload, retain=True)
    print(payload)

async def sensor_ds18b20_task(mqttconf):
    print("sensor_ds18b20_task started.")
    client = mqttconf.make_client(id_prefix="pico-sensor-ds-")
    client.connect()
    
    if not ds_roms:
        print("No DS18B20 sensor found!")
        return

    while True:
        led.on()
        try:
            ds_sensor.convert_temp()
            await asyncio.sleep(0.75)
            temperature = ds_sensor.read_temp(ds_roms[0])
            publish_data(client, mqttconf.topic, temperature, None)
        except Exception as e:
            print("Error reading sensor:", e)
        await asyncio.sleep(0.2)
        led.off()
        await asyncio.sleep(14.05)

async def sensor_dht11_task(mqttconf):
    print("sensor_dht11_task started.")
    client = mqttconf.make_client(id_prefix="pico-sensor-dht-")
    client.connect()
    while True:
        led.on()
        try:
            sensor.measure()
            temperature = sensor.temperature()
            humidity = sensor.humidity()
            publish_data(client, mqttconf.topic, temperature, humidity)
        except Exception as e:
            print("Error reading sensor:", e)
        await asyncio.sleep(0.2)
        led.off()
        await asyncio.sleep(14.8)

async def sensor_task(mqttconf):
    print("sensor_internal_task", mqttconf)
    client = mqttconf.make_client(id_prefix="pico-sensor-internal-")
    client.connect()
    while True:
        reading = internal_temp_sensor.read_u16() * conversion_factor
        temperature = 27 - (reading - 0.706) / 0.001721
        publish_data(client, mqttconf.topic + "/internal", temperature, None)
        await asyncio.sleep(15)

# Main function
async def main():
    wifi = MicroWifi()
    await wifi.connect()

    while True:
        mqttconf = MicroMqttConf()
        sensor_task1_obj = asyncio.create_task(sensor_task(mqttconf))
        sensor_task2_obj = asyncio.create_task(sensor_dht11_task(mqttconf))
        web_server_task  = asyncio.create_task(mqttconf.serveConfPage())

        await web_server_task
        await sensor_task1_obj
        await sensor_task2_obj
    

if __name__ == "__main__":
    asyncio.run(main())

