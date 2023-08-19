import aiohttp_jinja2
from aiohttp import web

@aiohttp_jinja2.template('index.html')
async def index(request):
    context = {
        "thermostats": {
            "01": {
                "name": "Living Room",
                "temperature": 22,
                "humidity": 40,
                "setpoint": 21,
                "mode": "heating"
            },
            "02": {
                "name": "Bedroom",
                "temperature": 20,
                "humidity": 45,
                "setpoint": 19,
                "mode": "cooling"
            }
        }
    }
    return context
    # return request.app['db']


@aiohttp_jinja2.template('thermostat.html')
async def thermostat_view(request):
    id = request.match_info['id']
    if not id in request.app["db"]["thermostats"]:
        raise web.HTTPNotFound(text="thermostat id not found")
    return  {
        "thermostat": request.app["db"]["thermostats"][id],
        "actuators": request.app["db"]["actuators"],
        "sensors": request.app["db"]["sensors"],
    }

def create_error_handler_middleware():

    async def handle404(request):
        return aiohttp_jinja2.render_template('404.html', request, {}, status=404)

    async def handle500(request):
        return aiohttp_jinja2.render_template('500.html', request, {}, status=500)

    client_error_routes = {
        404: handle404,
    }

    @web.middleware
    async def error_middleware(request, handler):
        try:
            return await handler(request)
        except web.HTTPException as ex:
            print(ex)
            return await client_error_routes[ex.status](request)
        except Exception:
            request.protocol.logger.exception("Error handling request")
            return await handle500(request)

    return error_middleware
