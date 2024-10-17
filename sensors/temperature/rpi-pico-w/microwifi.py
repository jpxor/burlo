# this module uses micropython to allow a user to connect the microcontroller to their wifi network
# requirements:
#  - try to connect to wifi using ssid and password stored in conf file
#  - if no file or fails to connect to wifi after x time, create Access Point (AP) and use microweb to serve page that lets a user
#    connect and set the ssid and password via http post to /update
#  - after update, reboot the microcontroller and connect to wifi

import asyncio
import network
import time
import json
import machine
import tinyweb
from tinyweb.server import urldecode_plus
from machine import Pin

led = Pin("LED", Pin.OUT)

class MicroWifi:
    def __init__(self, config_file='wifi_config.json', max_retries=3, retry_delay=5):
        led.off()
        self.config_file = config_file
        self.max_retries = max_retries
        self.retry_delay = retry_delay

    async def connect(self):
        self.wlan = network.WLAN(network.STA_IF)
        self.wlan.active(True)
        if self._load_config():
            for _ in range(self.max_retries):
                try:
                    self.wlan.connect(self.ssid, self.password)
                    for _ in range(10):  # Wait up to 10 seconds for connection
                        if self.wlan.isconnected():
                            print(f"Connected to {self.ssid}")
                            return True
                        await asyncio.sleep(1)
                except Exception as e:
                    print(f"Connection failed: {e}")
                await asyncio.sleep(self.retry_delay)
        
        print("Failed to connect, starting AP mode")
        self.wlan.active(False)
        await self._start_ap()

    def _load_config(self):
        try:
            with open(self.config_file, 'r') as f:
                config = json.load(f)
                self.ssid = config['ssid']
                self.password = config['password']
            return True
        except:
            print("No valid configuration found")
            return False

    def _save_config(self, ssid, password):
        with open(self.config_file, 'w') as f:
            json.dump({'ssid': ssid, 'password': password}, f)

    async def _start_ap(self):
        async def blink_led():
            while True:
                led.toggle()
                await asyncio.sleep(0.5)
        asyncio.create_task(blink_led())

        self.ap = network.WLAN(network.AP_IF)
        self.ap.config(essid='MicroAP', password='setupwifi')
        self.ap.active(True)
        print('Access Point IP:', self.ap.ifconfig()[0])

        while self.ap.active() == False:
            pass

        app = tinyweb.webserver()

        @app.route('/')
        async def index(request, response):
            await response.start_html()
            await response.start_html()
            wifi_setup_page = f"""
            <html>
            <head>
                <title>WiFi Setup</title>
                <style>
                    .container {{
                        display: flex;
                        justify-content: center;
                        align-items: center;
                        height: 100vh;
                        font-family: Arial, sans-serif;
                    }}
                    .form-wrapper {{
                        background-color: #f0f0f0;
                        padding: 2rem;
                        border-radius: 8px;
                        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
                    }}
                    .form-title {{
                        text-align: center;
                        color: #333;
                        margin-bottom: 1.5rem;
                    }}
                    .form-group {{
                        margin-bottom: 1rem;
                    }}
                    .form-label {{
                        display: block;
                        margin-bottom: 0.5rem;
                        color: #555;
                    }}
                    .form-input {{
                        width: 100%;
                        padding: 0.5rem;
                        border: 1px solid #ccc;
                        border-radius: 4px;
                    }}
                    .form-submit {{
                        width: 100%;
                        padding: 0.75rem;
                        background-color: #4CAF50;
                        color: white;
                        border: none;
                        border-radius: 4px;
                        cursor: pointer;
                    }}
                    .form-submit:hover {{
                        background-color: #45a049;
                    }}
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="form-wrapper">
                        <h1 class="form-title">WiFi Setup</h1>
                        <form action="/update" method="post">
                            <div class="form-group">
                                <label class="form-label" for="ssid">SSID:</label>
                                <input class="form-input" type="text" id="ssid" name="ssid" value="{self.ssid}">
                            </div>
                            <div class="form-group">
                                <label class="form-label" for="password">Password:</label>
                                <input class="form-input" type="password" id="password" name="password" value="{self.password}">
                            </div>
                            <input class="form-submit" type="submit" value="Submit">
                        </form>
                    </div>
                </div>
            </body>
            </html>
            """
            await response.send(wifi_setup_page)


        @app.route('/update', methods=['POST'])
        async def update(request, response):
            body = await request.read_body()
            form_data = dict(x.split('=') for x in body.split('&'))
            if form_data is None:
                await response.error(400, msg="no form data")
                return
    
            ssid = urldecode_plus(form_data.get('ssid'))
            password = urldecode_plus(form_data.get('password'))
    
            if ssid and password:
                self._save_config(ssid, password)
                self._load_config()
                await response.start_html()
                await response.send("ok")
            else:
                await response.error(400, msg="missing form data")

        print("Starting AP web server...")
        await app.run(host='0.0.0.0', port=80)

