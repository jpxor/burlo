import json
import paho.mqtt.client as mqtt
from threading import Lock

activeSensors = []
sensorMutex = Lock()


class SensorInterface():

    def subscribe(self, callback):
        if not hasattr(self, "callbacks"):
            self.callbacks = []
        self.callbacks.append(callback)

    def publish(self, data):
        print("publishing")
        if hasattr(self, "callbacks"):
            for cb in self.callbacks:
                cb(data)

    def stop(self):
        pass


class MqttSensor(SensorInterface):

    def __init__(self, topic, broker, port=1883):
        self.topic = topic
        self.client = mqtt.Client()
        self.client.on_connect = self.mqtt_on_connect
        self.client.on_message = self.mqtt_on_message
        self.client.on_disconnect = self.mqtt_on_disconnect
        self.client.username_pw_set("hvac", "hvac_pass")
        self.client.connect(broker, port, 60)
        with sensorMutex:
            self.client.loop_start()
            activeSensors.append(self)

    def stop(self):
        self.client.loop_stop()

    def mqtt_on_connect(self, client, opaque, flags, rc):
        print(f'connect rc={rc}, subscribing to "{self.topic}"')
        client.subscribe(self.topic)

    def mqtt_on_disconnect(self, client, opaque, rc):
        print(f'Disconnected with result code {str(rc)}, topic "{self.topic}"')

    def mqtt_on_message(self, client, opaque, msg):
        print(f'recieved message from topic "{msg.topic}"')
        try:
            data = json.loads(msg.payload)
            if "dewpoint" not in data and "temperature" in data and "humidity" in data:
                if data["humidity"] > 50:
                    t = data["temperature"]
                    rh = data["humidity"]
                    data["dewpoint"] = t - (100-rh) / 5
            self.publish(data)
        except:
            print(f'"mqtt sensor exception:"\
                  "broker "{self.broker}", topic "{self.topic}", payload {msg.payload}')


async def aiohttp_sensor_context(app):
    yield
    try:
        with sensorMutex:
            for client in activeSensors:
                client.stop()
    except Exception as e:
        print(e)


if __name__ == "__main__":
    print("testing mqtt enabled sensors")
    s = MqttSensor(topic="zigbee2mqtt/thermostats/01/office",
                   broker="192.168.50.193")

    s.subscribe(callback=lambda data: print(data))
    s.publish(data={"test": "pubsub"})

    import time
    time.sleep(2)

    input("Press Enter to stop...")
    print("done")
