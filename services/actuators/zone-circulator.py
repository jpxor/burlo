from Phidget22.PhidgetException import *
from Phidget22.Phidget import *
from Phidget22.Devices.DigitalOutput import *
import traceback
import time


def onAttach(self):
	print("Attach!")
	print(self)



def onDetach(self):
	print("Detach!")
	print(self)



def onError(self, code, description):
	print("Code: " + ErrorEventCode.getName(code))
	print("Description: " + str(description))
	print("----------")



def enable():
	try:
		digitalOutput0 = DigitalOutput()
		digitalOutput0.setOnAttachHandler(onAttach)
		digitalOutput0.setOnDetachHandler(onDetach)
		digitalOutput0.setOnErrorHandler(onError)

        # specify hub,port,etc
		digitalOutput0.openWaitForAttachment(5000)
		digitalOutput0.setDutyCycle(1)

	except PhidgetException as ex:
		traceback.print_exc()
		print("")
		print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



def disable():
	try:
		digitalOutput0 = DigitalOutput()
		digitalOutput0.setOnAttachHandler(onAttach)
		digitalOutput0.setOnDetachHandler(onDetach)
		digitalOutput0.setOnErrorHandler(onError)

        # specify hub,port,etc
		digitalOutput0.openWaitForAttachment(5000)
		digitalOutput0.setDutyCycle(0)

	except PhidgetException as ex:
		traceback.print_exc()
		print("")
		print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)



def test():
	try:
		digitalOutput0 = DigitalOutput()
		digitalOutput0.setOnAttachHandler(onAttach)
		digitalOutput0.setOnDetachHandler(onDetach)
		digitalOutput0.setOnErrorHandler(onError)

        # specify hub,port,etc
		digitalOutput0.openWaitForAttachment(5000)
		digitalOutput0.setDutyCycle(1)
		time.sleep(5)
		digitalOutput0.close()

	except PhidgetException as ex:
		traceback.print_exc()
		print("")
		print("PhidgetException " + str(ex.code) + " (" + ex.description + "): " + ex.details)


# get the args
if len(sys.argv) != 2:
    print("Usage: zone-circulator.py [enable|disable|test]")
    sys.exit(1)

cmd = sys.argv[1]

if cmd == "enable":
    enable()
elif cmd == "disable":  
    disable()
elif cmd == "test":
    test()
else:
    print("Invalid command. Must be enable, disable or test.")
    sys.exit(1)

