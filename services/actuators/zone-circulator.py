from Phidget22.PhidgetException import *
from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import *
import traceback
import time


def onAttach(self):
    print(f"Attached {self}")

def onDetach(self):
    print(f"Detached {self}")

def onError(self, code, description):
    print("Code: " + ErrorEventCode.getName(code))
    print("Description: " + str(description))
    print("----------")



def enable():
    try:
        digitalOutput0 = DigitalOutput()
        digitalOutput0.setChannel(0)
        digitalOutput0.setHubPort(0)

        digitalOutput0.setOnAttachHandler(onAttach)
        digitalOutput0.setOnDetachHandler(onDetach)
        digitalOutput0.setOnErrorHandler(onError)

        digitalOutput0.openWaitForAttachment(5000)
        digitalOutput0.setDutyCycle(1)
        digitalOutput0.close()

    except PhidgetException as ex:
        traceback.print_exc()
        print("")
        print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



def disable():
    try:
        digitalOutput0 = DigitalOutput()
        digitalOutput0.setChannel(0)
        digitalOutput0.setHubPort(0)

        digitalOutput0.setOnAttachHandler(onAttach)
        digitalOutput0.setOnDetachHandler(onDetach)
        digitalOutput0.setOnErrorHandler(onError)

        digitalOutput0.openWaitForAttachment(5000)
        digitalOutput0.setDutyCycle(0)
        digitalOutput0.close()

    except PhidgetException as ex:
        traceback.print_exc()
        print("")
        print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



def test():
    try:
        digitalOutput0 = DigitalOutput()
        digitalOutput0.setChannel(0)
        digitalOutput0.setHubPort(0)

        digitalOutput0.setOnAttachHandler(onAttach)
        digitalOutput0.setOnDetachHandler(onDetach)
        digitalOutput0.setOnErrorHandler(onError)

        digitalOutput0.openWaitForAttachment(5000)
        digitalOutput0.setDutyCycle(1)
        time.sleep(5)
        digitalOutput0.setDutyCycle(0)
        digitalOutput0.close()

    except PhidgetException as ex:
        traceback.print_exc()
        print("")
        print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



def query():
    try:
        do = DigitalOutput()
        do.setChannel(0)
        do.setHubPort(0)

        do.setOnAttachHandler(onAttach)
        do.setOnDetachHandler(onDetach)
        do.setOnErrorHandler(onError)

        do.openWaitForAttachment(5000)
        print(f"getDutyCycle {do.getDutyCycle()}")
        print(f"getChannel {do.getChannel()}")
        print(f"getHubPort {do.getHubPort()}")
        print(f"getHub {do.getHub()}")
        print(f"getDeviceSerialNumber {do.getDeviceSerialNumber()}")
        do.close()

    except PhidgetException as ex:
        traceback.print_exc()
        print("")
        print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



# get the args
if len(sys.argv) != 2:
    print("Usage: zone-circulator.py [enable|disable|test|query]")
    sys.exit(1)

cmd = sys.argv[1]

if cmd == "enable":
    enable()
elif cmd == "disable":  
    disable()
elif cmd == "test":
    test()
elif cmd == "query":
    query()
else:
    print("Invalid command. Must be enable, disable or test.")
    sys.exit(1)

