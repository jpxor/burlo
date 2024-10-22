import dht
import machine
import asyncio

from microwifi import MicroWifi
from micromqttconf import MicroMqttConf
from machine import Pin

# internal LED
led = Pin("LED", Pin.OUT)

# Initialize DHT22 sensor
sensor = dht.DHT22(machine.Pin(28))

# Publish data to MQTT server
def publish_data(client, topic, temperature, humidity):
    payload = f"{{\"temperature\": {temperature}, \"humidity\": {humidity}}}"
    client.publish(topic, payload, retain=True)
    print(payload)


async def sensor_dht22_task(mqttconf):
    print("sensor_dht22_task started.")
    client = mqttconf.make_client(id_prefix="pico-sensor-dht-")
    client.connect()
    while True:
        await asyncio.sleep(2)
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
        await asyncio.sleep(12.8)


# Main function
async def main():
    wifi = MicroWifi()
    await wifi.connect()

    while True:
        mqttconf = MicroMqttConf()
        sensor_task_obj = asyncio.create_task(sensor_dht22_task(mqttconf))
        web_server_task  = asyncio.create_task(mqttconf.serveConfPage())
        await asyncio.gather(web_server_task, sensor_task_obj)
    

if __name__ == "__main__":
    asyncio.run(main())
