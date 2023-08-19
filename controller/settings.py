import pathlib
import yaml
import json
import os

dbpath = './config/controller.db'

def load_config(argv):
    if len(argv) == 1:
        path = "./config/configuration.yaml"
    elif os.path.exists(argv[1]):
        path = argv[1]
    else:
        print(f'FATAL: config file not found: {argv[1]}')
        exit(-1)
    with open(path) as f:
        return yaml.safe_load(f)

def load_db():
    if os.path.exists(dbpath):
        with open(dbpath, 'r') as f:
            return json.load(f)
    return {
        "thermostats": {},
        "actuators": [],
        "sensors": [],
    }

def save_db(data):
    with open(dbpath, 'w') as f:
        json.dump(data, f)
