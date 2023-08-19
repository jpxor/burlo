import sys
import logging

import jinja2
import aiohttp_jinja2

from aiohttp import web
from views import index,  thermostat_view, create_error_handler_middleware
from settings import load_config, load_db, save_db
from threading import Lock


async def on_shutdown(app):
    save_db(app['db'])

async def phigets_context(app):
    # find and load configured phiget devices
    yield
    # close all phiget devices

def main():
    logging.basicConfig(level=logging.INFO)
    app = web.Application()
    app.add_routes(
        [
            web.get('/', index),
            web.get('/thermostat/{id}', thermostat_view),
            web.static('/assets/', "./www/static"),
        ]
    )
    app.middlewares.append(create_error_handler_middleware())

    app['mutex'] = Lock()
    app['config'] = load_config(sys.argv)
    app['db'] = load_db()

    print(app['db'])

    app.on_shutdown.append(on_shutdown)
    app.cleanup_ctx.append(phigets_context)

    aiohttp_jinja2.setup(app, loader=jinja2.FileSystemLoader('./www/templates'))
    web.run_app(app, host='0.0.0.0', port=8000, access_log_format=" :: %r %s %t")


if __name__ == '__main__':
    main()