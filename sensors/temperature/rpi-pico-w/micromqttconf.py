import json
import machine
import ubinascii

import tinyweb
from tinyweb.server import urldecode_plus

from umqttsimple import MQTTClient

class MicroMqttConf:
    def __init__(self, config_file='mqtt_config.json'):
        self.config_file = config_file
        self.host = "localhost:1883"
        self.user = ""
        self.password = ""
        self.topic = "mqttconf/default/topic"
        self._load_config()

    def make_client(self, id_prefix=""):
        client_id = id_prefix+str(ubinascii.hexlify(machine.unique_id()))
        server, port = self.host.split(":", 1)
        return MQTTClient(client_id, server, port=int(port), user=self.user, password=self.password, keepalive=60)

    def _load_config(self):
        print("mqttconf: loading config")
        try:
            with open(self.config_file, 'r') as f:
                config = json.load(f)
                self.host = config['host']
                self.user = config['user']
                self.password = config['password']
                self.topic = config['topic']
            return True
        except:
            print("No valid configuration found")
            return False

    def _save_config(self, host, user, password, topic):
        print("mqttconf: saving config")
        with open(self.config_file, 'w') as f:
            json.dump({'host': host, 'user': user, 'password': password, 'topic': topic}, f)

    async def serveConfPage(self):
        await self.start_web_app()

    async def start_web_app(self):
        app = tinyweb.webserver()
        
        @app.route('/')
        async def index(request, response):
            mqtt_conf_page = f"""
            <html>
            <head>
                <title>MQTT Setup</title>
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
                        <h1 class="form-title">MQTT Setup</h1>
                        <form action="/update" method="post">
                            <div class="form-group">
                                <label class="form-label" for="host">Host (ip:port):</label>
                                <input class="form-input" type="text" id="host" name="host" value="{self.host}">
                            </div>
                            <div class="form-group">
                                <label class="form-label" for="user">User:</label>
                                <input class="form-input" type="text" id="user" name="user" value="{self.user}">
                            </div>
                            <div class="form-group">
                                <label class="form-label" for="password">Password:</label>
                                <input class="form-input" type="password" id="password" name="password" value="{self.password}">
                            </div>
                            <div class="form-group">
                                <label class="form-label" for="topic">Topic:</label>
                                <input class="form-input" type="text" id="topic" name="topic" value="{self.topic}">
                            </div>
                            <input class="form-submit" type="submit" value="Submit">
                        </form>
                    </div>
                </div>
            </body>
            </html>
            """
            await response.start_html()
            await response.send(mqtt_conf_page)

        @app.route('/update', methods=['POST'])
        async def update(request, response):
            body = await request.read_body()
            form_data = dict(x.split('=') for x in body.split('&'))
            if form_data is None:
                await response.error(400, msg="no form data")
                return

            host = urldecode_plus(form_data.get('host'))
            user = urldecode_plus(form_data.get('user'))
            password = urldecode_plus(form_data.get('password'))
            topic = urldecode_plus(form_data.get('topic'))

            print("updating:", host, user, password, topic)
            if host and user and password and topic:
                self._save_config(host, user, password, topic)
                if self._load_config():
                    await response.start_html()
                    await response.send("success!")
                    return

            await response.error(400, msg="missing form data")

        print("Starting mqtt-config web server...")
        await app.run(host='0.0.0.0', port=80)
