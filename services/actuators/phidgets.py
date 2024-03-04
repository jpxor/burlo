from aiohttp import web

from Phidget22.Phidget import *
from Phidget22.PhidgetException import *
from Phidget22.Devices.DigitalOutput import *

import sys
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
        return {
            "name": self.name,
            "phidget": str(self.phidget),
            "state": self.phidget.getState(), # DigitalOutput only
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

        channel = data.get("channel", -1)
        if not isinstance(channel, int):
            return web.Response(status=400, text="channel must be an integer")

        hub_port = data.get("hub_port", -1)
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
    return web.Response(status=200)


async def detach_phidget_channel(request):
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
        phidget = named_phidgets[name].phidget
        await phidget.close()
    except Exception as ex:
        traceback.print_exc()
        raise web.HTTPBadRequest(reason=str(ex))


async def get_phidgets_state(request):
    serializables = [phiwrap.toSerializable() for phiwrap in named_phidgets.values()]
    return web.json_response(serializables)


app = web.Application()
app.add_routes([
    web.post('/phidgets/digital_out', set_digital_output),
    web.post('/phidgets/detach', detach_phidget_channel),
    web.get('/phidgets/state', get_phidgets_state),
])


def run():
    web.run_app(app, port=4000)


def test():
    async def test_phidgets():
        async with aiohttp.ClientSession() as session:
            # Test setting digital output
            data = {'name': 'my_device', 'channel': 0, 'hub_port': 0, 'target_state': 2}
            async with session.post('http://192.168.50.193:4000/phidgets/digital_out/', json=data) as resp:
                print(resp.status, resp.reason)
            
            # Test getting state
            async with session.get('http://192.168.50.193:4000/phidgets/state') as resp:
                print(await resp.json())
                
            # Test detaching
            data = {'name': 'my_device'}
            async with session.post('http://192.168.50.193:4000/phidgets/detach/', json=data) as resp:
                print(resp.status)
                
    loop = asyncio.get_event_loop()
    loop.run_until_complete(test_phidgets())



if len(sys.argv) != 2:
    print("Usage: phidgets.py [test|run]")
    sys.exit(1)
    
cmd = sys.argv[1]
if cmd == "test":
    test()
elif cmd == "run":
    run()
else:
    print("Invalid argument. Usage: phidgets.py [test|run]")

