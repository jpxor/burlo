package main

import (
	"burlo/pkg/models/controller"
	"fmt"

	"github.com/gorilla/websocket"
)

var ws = WebSocketConnections{
	connections: make(map[*websocket.Conn]Session),
}

func pushThermostatToDashboards(tstat controller.Thermostat) {
	fmt.Println("pushThermostatToDashboards:", tstat)
	// ws.writeAll(tstat)
}

func pushWeatherToDashboards(weather Weather) {
	fmt.Println("pushWeatherToDashboards:", weather)
	// ws.writeAll(weather)
}

func pushSetpointToDashboards(sp SetpointData) {
	var cmd = struct {
		Command string `json:"command"`
		Id      string `json:"id"`
		Html    string `json:"html"`
	}{
		Command: "setInnerHTML",
		Id:      "setpoint-value",
	}
	temp := sp.HeatingSetpoint
	if sp.Mode == Cool {
		temp = sp.CoolingSetpoint
	}
	ws.forEach(func(conn *websocket.Conn, sess Session) {
		cmd.Html = fmt.Sprintf("%.0f", temp.asFloat(sess.Unit))
		conn.WriteJSON(cmd)
	})
}
