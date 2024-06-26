from aiohttp import web

from Phidget22.Phidget import *
from Phidget22.PhidgetException import *
from Phidget22.Devices.DigitalOutput import *
from Phidget22.Devices.VoltageOutput import *

import sys
import json
import aiohttp
import asyncio
import traceback

## active phidget channels
named_phidgets = {}

class NamedPhidget:
    def __init__(self, name, phidget):
        self.phidget = phidget
        self.name = name
    
    def toSerializable(self):
        if isinstance(self.phidget, DigitalOutput):
            return {
                "name": self.name,
                "phidget": str(self.phidget),
                "state": self.phidget.getState(),
            }
        if isinstance(self.phidget, VoltageOutput): 
            return {
                "name": self.name,
                "phidget": str(self.phidget),
                "voltage": self.phidget.getVoltage(),
            }


def name_from_phidget(phidget):
    for name, wrapper in named_phidgets.items():
        if phidget == wrapper.phidget:
            return name


def onAttach(self):
    print(f"Attached: {self}")


def onDetach(self):
    print(f"Detached: {self}")
    del named_phidgets[name_from_phidget(self)]


def onError(self, code, description):
    print("Device: " + str(self))
    print("Code: " + ErrorEventCode.getName(code))
    print("Description: " + str(description))
    print("----------")


async def set_digital_output(request):
    try:
        data = await request.json()

        if "name" not in data or "target_state" not in data:
            return web.Response(status=400, text="requires name (str) and target_state (bool)")
        
        name = data["name"]
        if not isinstance(name, str):
            return web.Response(status=400, text="name must be a string")

        target_state = data["target_state"]
        if not isinstance(target_state, bool):
            return web.Response(status=400, text="target_state must be a boolean")

        channel = data.get("channel", -2)
        if not isinstance(channel, int):
            return web.Response(status=400, text="channel must be an integer")

        hub_port = data.get("hub_port", -2)
        if not isinstance(hub_port, int):
            return web.Response(status=400, text="hub_port must be an integer")
    
    except:
        return web.Response(status=400, text="bad request")

    phiwrap = named_phidgets.get(name)
    if not phiwrap:
        if channel == -1 or hub_port == -1:
            return web.Response(status=400, text="name not found; channel and hub_port must be set")
        try:
            do = DigitalOutput()
            do.setChannel(channel)
            do.setHubPort(hub_port)
            
            do.setOnAttachHandler(onAttach)
            do.setOnDetachHandler(onDetach)
            do.setOnErrorHandler(onError)
            do.openWaitForAttachment(5000)

            phiwrap = NamedPhidget(name, do)
            named_phidgets[name] = phiwrap

        except PhidgetException as ex:
            traceback.print_exc()
            return web.Response(status=500, text=str(ex))

    phiwrap.phidget.setState(target_state)
    return web.Response(status=200, text="ACK")


async def set_voltage_output(request):
    try:
        data = await request.json()

        if "name" not in data or "target_state" not in data:
            return web.Response(status=400, text="requires name (str) and target_state (float)")
        
        name = data["name"]
        if not isinstance(name, str):
            return web.Response(status=400, text="name must be a string")

        target_state = data["target_state"]
        if type(target_state) not in (int, float):
            return web.Response(status=400, text="target_state must be an int or float")

        if target_state > 10.0 or target_state < -10.0:
            return web.Response(status=400, text="target_state must be +/- 10V")

        channel = data.get("channel", -2)
        if not isinstance(channel, int):
            return web.Response(status=400, text="channel must be an integer")

        hub_port = data.get("hub_port", -2)
        if not isinstance(hub_port, int):
            return web.Response(status=400, text="hub_port must be an integer")
    
    except:
        return web.Response(status=400, text="bad request")

    phiwrap = named_phidgets.get(name)
    if not phiwrap:
        if channel == -1 or hub_port == -1:
            return web.Response(status=400, text="name not found; channel and hub_port must be set")
        try:
            vo = VoltageOutput()
            vo.setChannel(channel)
            vo.setHubPort(hub_port)
            
            vo.setOnAttachHandler(onAttach)
            vo.setOnDetachHandler(onDetach)
            vo.setOnErrorHandler(onError)
            vo.openWaitForAttachment(5000)

            phiwrap = NamedPhidget(name, vo)
            named_phidgets[name] = phiwrap

        except PhidgetException as ex:
            traceback.print_exc()
            return web.Response(status=500, text=str(ex))

    phiwrap.phidget.setVoltage(target_state)
    return web.Response(status=200, text="ACK")


async def close_phidget_channel(request):
    try:
        data = await request.json()

        if "name" not in data:
            return web.Response(status=400, text="requires name (str)")

        name = data["name"]
        if not isinstance(name, str):
            return web.Response(status=400, text="name must be a string")

    except:
        return web.Response(status=400, text="bad request")

    if name not in named_phidgets:
        return web.Response(status=400, text="no phidget by that name")

    try:
        named_phidgets[name].phidget.close()
        del named_phidgets[name]
        return web.Response(status=200, text="ACK")

    except Exception as ex:
        traceback.print_exc()
        raise web.HTTPBadRequest(reason=str(ex))


async def get_phidgets_state(request):
    serializables = [phiwrap.toSerializable() for phiwrap in named_phidgets.values()]
    out = """
<!DOCTYPE html>
<html>
<head>
    <title>Phidgets State</title>
</head>
<body>
    <p>[OK] /services/actuators/phidgets<p>
    <pre>
""" + json.dumps(serializables, indent=4) + """
    </pre>
    <script>
        async function sendDO(val) {
            const url = '/phidgets/digital_out';
            const payload = {
                'name': 'ZoneCirculator',
                'channel': 0,
                'hub_port': 0,
                'target_state': val
            };
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });
            const data = await response.text();
            location.reload();
        }
        async function sendVO(val) {
            const url = '/phidgets/voltage_out';
            const payload = {
                'name': 'Dewpoint',
                'channel': 0,
                'hub_port': 1,
                'target_state': val
            };
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });
            const data = await response.text();
            location.reload();
        }
    </script>
    <p>Testing:</p>
    <button onclick="sendDO(true)">digital_out set true</button>
    <button onclick="sendDO(false)">digital_out set false</button>
    <br>
    <button onclick="sendVO(0)">voltage_out set 0.0V</button>
    <button onclick="sendVO(5.0)">voltage_out set 5.0V</button>
    <button onclick="sendVO(10.0)">voltage_out set 10.0V</button>
</body>
</html>"""
    return web.Response(status=200, text=out, content_type="text/html")


app = web.Application()
app.add_routes([
    web.post('/phidgets/digital_out', set_digital_output),
    web.post('/phidgets/voltage_out', set_voltage_output),
    web.post('/phidgets/close', close_phidget_channel),
    web.get('/phidgets/state', get_phidgets_state),
])


def run():
    web.run_app(app, port=4002)


def sanity_test(): # happy paths only
    async def test_phidgets():
        async with aiohttp.ClientSession() as session:
            # Test setting digital output
            data = {'name': 'my_device', 'channel': 0, 'hub_port': 0, 'target_state': True}
            async with session.post('http://192.168.50.193:4000/phidgets/digital_out', json=data) as resp:
                if resp.status != 200:
                    print(resp.status, resp.reason)
            
            # Test getting state
            async with session.get('http://192.168.50.193:4000/phidgets/state') as resp:
                print(await resp.text())

            # Test setting digital output
            data = {'name': 'my_device', 'target_state': False}
            async with session.post('http://192.168.50.193:4000/phidgets/digital_out', json=data) as resp:
                if resp.status != 200:
                    print(resp.status, resp.reason)

            # Test getting state
            async with session.get('http://192.168.50.193:4000/phidgets/state') as resp:
                print(await resp.text())
                
            # Test closing
            data = {'name': 'my_device'}
            async with session.post('http://192.168.50.193:4000/phidgets/close', json=data) as resp:
                print("closing:", resp.status, resp.reason)

    loop = asyncio.get_event_loop()
    loop.run_until_complete(test_phidgets())



if len(sys.argv) != 2:
    print("Usage: phidgets.py [test|run]")
    sys.exit(1)
    
cmd = sys.argv[1]
if cmd == "test":
    sanity_test()
elif cmd == "run":
    run()
else:
    print("Invalid argument. Usage: phidgets.py [test|run]")
