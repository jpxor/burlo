from threading import Lock

from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import DigitalOutput
from Phidget22.Devices.Manager import *

supported_channel_classes = [
    ChannelClass.PHIDCHCLASS_DIGITALOUTPUT,
    # ChannelClass.PHIDCHCLASS_VOLTAGEOUTPUT,
]

attachedChannels = []
in_use_channels = []
phidgetMutex = Lock()


class ActuatorInterface:
    def set_state(self, state):
        raise NotImplementedError


class PhidgetsDigitalOutput(ActuatorInterface):
    def __init__(self, phidget_device):
        self.phidget_device = phidget_device
        self.digital_output = DigitalOutput()
        self.digital_output.setDeviceSerialNumber(
            phidget_device.getDeviceSerialNumber())
        self.digital_output.setHubPort(phidget_device.getHubPort())
        self.digital_output.setChannel(phidget_device.getChannel())
        self.digital_output.openWaitForAttachment(5000)

    def __del__(self):
        self.digital_output.close()
        with phidgetMutex:
            in_use_channels.remove(self.phidget_device)

    def set_state(self, state):
        self.digital_output.setState(state)


def get_actuators_for_render():
    actuators = []
    with phidgetMutex:
        for chan in attachedChannels:
            actuators.append({
                "devname": chan.getDeviceName(),
                "chaname": chan.getChannelName(),
                "hubserial": chan.getDeviceSerialNumber(),
                "hubport": chan.getHubPort(),
                "chid": chan.getChannel(),
                "in-use": chan in in_use_channels,
            })
    return actuators


def phidget_channel_match(chan_a, chan_b):
    if chan_a.getChannel() == chan_b.getChannel():
        if chan_a.getHubPort() == chan_b.getHubPort():
            if chan_a.getDeviceSerialNumber() == chan_b.getDeviceSerialNumber():
                return True
    return False


def create_actuator(selected_chan):
    with phidgetMutex:
        for chan in attachedChannels:
            if phidget_channel_match(chan, selected_chan):
                if chan.getChannelClass() == ChannelClass.PHIDCHCLASS_DIGITALOUTPUT:
                    in_use_channels.append(chan)
                    return PhidgetsDigitalOutput(chan)
                else:
                    print("unsupported channel class", chan.getChannelClass())
                    break
    return None


def ManagerOnAttach(self, device):
    with phidgetMutex:
        if device.getIsChannel():
            if device.getChannelClass() in supported_channel_classes:
                attachedChannels.append(device)


def ManagerOnDetach(self, device):
    with phidgetMutex:
        attachedChannels.remove(device)


def print_phidget_exeption(e):
    print("Phidget Exception: " + str(e.code) + " - " + str(e.details))


def open_phidget_manager():
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
        print_phidget_exeption(e)
        exit(1)


async def aiohttp_phidget_context(app):
    manager = open_phidget_manager()
    yield
    try:
        with phidgetMutex:
            for chan in attachedChannels:
                chan.close()
        manager.close()
    except PhidgetException as e:
        print_phidget_exeption(e)


if __name__ == "__main__":
    print("testing phidgets")
    manager = open_phidget_manager()

    import time
    time.sleep(2)

    with phidgetMutex:
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
        print_phidget_exeption(e)
    print("done")
