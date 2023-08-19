from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import DigitalOutput

from Phidget22.Devices.Manager import *

def onAttach(self,sec):
    device = self
    deviceName = device.getDeviceName()
    serialNumber = device.getDeviceSerialNumber()
    chid = device.getChannel()
    print("second param:", sec)
    print("Hello to Device " + str(deviceName) + ", Serial Number: " + str(serialNumber) + ", channel id: " + str(chid))

def onDetach(self):
    device = self
    deviceName = device.getDeviceName()
    serialNumber = device.getDeviceSerialNumber()
    chid = device.getChannel()
    print("Goodbye Device " + str(deviceName) + ", Serial Number: " + str(serialNumber) + ", channel id: " + str(chid))

def onError(self, code, description):
	print("Code [" + str(self.getChannel()) + "]: " + ErrorEventCode.getName(code))
	print("Description [" + str(self.getChannel()) + "]: " + str(description))
	print("----------")

def init_actuators():
     pass

while True:
    dout = DigitalOutput()
    dout.setOnAttachHandler(onAttach)
    dout.setOnDetachHandler(onDetach)
    try: dout.openWaitForAttachment(1000)
    except PhidgetException as e:
        print(e)
        break