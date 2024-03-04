from aiohttp import web

from Phidget22.Phidget import *
from Phidget22.PhidgetException import *
from Phidget22.Devices.DigitalOutput import *

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
    
    def toJSON(self):
        return json.dumps(self, default=lambda o: o.__dict__, sort_keys=True, indent=4)


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
    data = await request.json()
    print("set_digital_output", data)
    
    name = data['name']
    if not isinstance(name, str):
        return web.Response(status=400, text="name must be a string")

    channel = data['channel']
    if not isinstance(channel, int):
        return web.Response(status=400, text="channel must be an integer")

    hub_port = data['hub_port'] 
    if not isinstance(hub_port, int):
        return web.Response(status=400, text="hub_port must be an integer")

    target_state = data['target_state']
    if not isinstance(target_state, bool):
        return web.Response(status=400, text="target_state must be a boolean")
    
    phiwrap = named_phidgets.get(name)
    if not phiwrap:
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
    data = await request.json()
    print("detach_phidget_channel", data)
    
    name = data['name']
    if not isinstance(name, str):
        return web.Response(status=400, text="name must be a string")

    try:
        phidget = named_phidgets[name].phidget
        await phidget.close()
    except Exception as ex:
        traceback.print_exc()
        raise web.HTTPBadRequest(reason=str(ex))


async def get_phidgets_state(request):
    return web.json_response(named_phidgets)


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

