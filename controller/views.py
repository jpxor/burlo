import traceback
import aiohttp_jinja2

from aiohttp import web
from actuators import get_actuators_for_render
from sensors import get_sensors_for_render


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


def create_error_handler_middleware():

    async def handle404(request):
        return aiohttp_jinja2.render_template('404.html', request, {}, status=404)

    async def handle500(request, context):
        return aiohttp_jinja2.render_template('500.html', request, context, status=500)

    client_error_routes = {
        404: handle404,
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
