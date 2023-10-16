import traceback
import aiohttp_jinja2

from aiohttp import web
from actuators import get_actuators_for_render
from sensors import MqttSensor, get_sensors_for_render
from settings import save_db


@aiohttp_jinja2.template('index.html')
async def index(request):
    context = {
        "thermostats": {
            "01": {
                "name": "Office",
                "temperature": 22,
                "humidity": 40,
                "setpoint": 21,
                "mode": "heating"
            },
            "02": {
                "name": "GuestRoom",
                "temperature": 20,
                "humidity": 45,
                "setpoint": 19,
                "mode": "cooling"
            }
        }
    }
    return context
    # return request.app['db']


@aiohttp_jinja2.template('actuators.html')
async def actuators_view(request):
    return {
        "actuators": get_actuators_for_render()
    }


@aiohttp_jinja2.template('sensors.html')
async def sensors_view(request):
    return {
        "sensors": get_sensors_for_render()
    }

async def post_sensor(request):
    data = await request.post()
    topic = data['mqttSubject']
    broker = data['brokerIpAddress']
    try:
        port = int(data['brokerPort'])
    except ValueError:
        raise web.HTTPBadRequest(text="Invalid port number")
    if 0 > port or port > 65535:
        raise web.HTTPBadRequest(text="Port number out of range")
    # create new sensor
    MqttSensor(topic, broker, port)
    # save state
    request.app['db']['sensors'].append({
        "type": "mqtt",
        "topic": topic,
        "broker": broker,
        "port": port,
    })
    save_db(request.app['db'])
    # redirect to reload sensors page
    return web.HTTPFound('/sensors')


@aiohttp_jinja2.template('thermostat.html')
async def thermostat_view(request):
    id = request.match_info['id']
    if not id in request.app["db"]["thermostats"]:
        raise web.HTTPNotFound(text="thermostat id not found")
    return {
        "thermostat": request.app["db"]["thermostats"][id],
        "actuators": get_actuators_for_render(),
        "sensors": get_sensors_for_render(),
    }


async def test_view(request):
    class TestEx(Exception):
        pass
    sts = request.match_info['status']
    if sts == "500":
        raise TestEx("testing 500")
    if sts == "404":
        raise web.HTTPNotFound(text="testing 404")
    if sts == "400":
        raise web.HTTPBadRequest(text="testing 400")


def create_error_handler_middleware():

    async def handle404(request, context={}):
        return aiohttp_jinja2.render_template('404.html', request, context, status=404)

    async def handle400(request, context={}):
        return aiohttp_jinja2.render_template('400.html', request, context, status=400)

    async def handle500(request, context):
        return aiohttp_jinja2.render_template('500.html', request, context, status=500)

    client_error_routes = {
        404: handle404,
        400: handle400,
    }

    @web.middleware
    async def error_middleware(request, handler):
        try:
            return await handler(request)
        except web.HTTPException as ex:
            if ex.status in client_error_routes:
                return await client_error_routes[ex.status](request)
            raise ex
        except Exception:
            request.protocol.logger.exception("Error handling request")
            return await handle500(request, {
                "backtrace": traceback.format_exc()
            })

    return error_middleware
