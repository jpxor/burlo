from threading import Lock

from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import DigitalOutput
from Phidget22.Devices.Manager import *

import Phidget22.ChannelClass
supported_channel_classes = [
    ChannelClass.PHIDCHCLASS_DIGITALOUTPUT,
    ChannelClass.PHIDCHCLASS_VOLTAGEOUTPUT,
]

attachedChannels = []
phigetMutex = Lock()


class ActuatorInterface:
    def set_state(self, state):
        raise NotImplementedError


class PhidgetsDigitalOutput(ActuatorInterface):
    def __init__(self, serialnumber, hubport, chid):
        self.serial_number = serialnumber
        self.hubport = hubport
        self.chid = chid
        self.digital_output = DigitalOutput()
        self.digital_output.setDeviceSerialNumber(serialnumber)
        self.digital_output.setHubPort(hubport)
        self.digital_output.setChannel(chid)
        self.digital_output.openWaitForAttachment(5000)

    def __del__(self):
        self.digital_output.close()

    def set_state(self, state):
        self.digital_output.setState(state)


def get_actuators_for_render():
    actuators = []
    with phigetMutex:
        for chan in attachedChannels:
            actuators.append({
                "name": chan.getDeviceName(),
                "hubserial": chan.getDeviceSerialNumber(),
                "hubport": chan.getHubPort(),
                "chid": chan.getChannel()
            })
    return actuators


def ManagerOnAttach(self, device):
    with phigetMutex:
        if device.getIsChannel():
            if device.getChannelClass() in supported_channel_classes:
                attachedChannels.append(device)


def ManagerOnDetach(self, device):
    with phigetMutex:
        attachedChannels.remove(device)


def print_phiget_exeption(e):
    print("Phidget Exception: " + str(e.code) + " - " + str(e.details))


def open_phiget_manager():
    try:
        manager = Manager()
    except RuntimeError as e:
        print("Runtime Error " + e.details)
        exit(1)
    try:
        manager.setOnAttachHandler(ManagerOnAttach)
        manager.setOnDetachHandler(ManagerOnDetach)
        manager.open()
        return manager
    except PhidgetException as e:
        print_phiget_exeption(e)
        exit(1)


async def aiohttp_phiget_context(app):
    manager = open_phiget_manager()
    yield
    try:
        with phigetMutex:
            for chan in attachedChannels:
                chan.close()
        manager.close()
    except PhidgetException as e:
        print_phiget_exeption(e)


if __name__ == "__main__":
    print("testing phigets")
    manager = open_phiget_manager()

    import time
    time.sleep(2)

    with phigetMutex:
        for chan in attachedChannels:
            name = chan.getDeviceName()
            sn = chan.getDeviceSerialNumber()
            chid = chan.getChannel()
            port = chan.getHubPort()
            chname = chan.getChannelName()
            print(f'name: {name}, chname: {chname}, S/N: {sn}, port: {port}, chid: {chid}')

    try:
        manager.close()
    except PhidgetException as e:
        print_phiget_exeption(e)
    print("done")
