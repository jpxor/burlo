from threading import Lock

from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import DigitalOutput
from Phidget22.Devices.Manager import *


attachedChannels = []
phigetMutex = Lock()


def ManagerOnAttach(self, device):
    with phigetMutex:
        attachedChannels.append(device)

    print("self is manager? " + str(self))
    deviceName = device.getDeviceName()
    serialNumber = device.getDeviceSerialNumber()
    chid = device.getChannel()
    print("Hello to Device " + str(deviceName) + ", Serial Number: " +
          str(serialNumber) + ", channel id: " + str(chid))


def ManagerOnDetach(self, device):
    with phigetMutex:
        attachedChannels.remove(device)

    deviceName = device.getDeviceName()
    serialNumber = device.getDeviceSerialNumber()
    chid = device.getChannel()
    print("Goodbye Device " + str(deviceName) + ", Serial Number: " +
          str(serialNumber) + ", channel id: " + str(chid))


def print_phiget_exeption(e):
    print("Phidget Exception: " + str(e.code) + " - " + str(e.details))


def open_phiget_manager():
    try:
        manager = Manager()
    except RuntimeError as e:
        print("Runtime Error " + e.details + ", Exiting...\n")
        exit(1)

    try:
        manager.setOnAttachHandler(ManagerOnAttach)
        manager.setOnDetachHandler(ManagerOnDetach)
    except PhidgetException as e:
        print_phiget_exeption(e)
        exit(1)

    try:
        manager.open()
    except PhidgetException as e:
        print_phiget_exeption(e)
        exit(1)

    return manager


def aiohttp_phiget_context(app):
    manager = open_phiget_manager()
    yield
    try:
        manager.close()
    except PhidgetException as e:
        print_phiget_exeption(e)


if __name__ == "__main__":
    print("testing phigets")
    manager = open_phiget_manager()

    import time
    time.sleep(2)

    with phigetMutex:
        print(attachedChannels)
        for chan in attachedChannels:
            name = chan.getDeviceName()
            sn = chan.getDeviceSerialNumber()
            chid = chan.getChannel()
            port = chan.getHubPort()
            print(chid, name, port, sn)

    try:
        manager.close()
    except PhidgetException as e:
        print_phiget_exeption(e)
    print("done")
