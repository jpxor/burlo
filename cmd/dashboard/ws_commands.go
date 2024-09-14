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
	// ws.writeAll(tstat)
}

func pushWeatherToDashboards(weather Weather) {
	// ws.writeAll(weather)
}

func pushSetpointToDashboards(temp Temperature) {
	var cmd = struct {
		Command string `json:"command"`
		Id      string `json:"id"`
		Html    string `json:"html"`
	}{
		Command: "setInnerHTML",
		Id:      "setpoint-value",
	}
	ws.forEach(func(conn *websocket.Conn, sess Session) {
		cmd.Html = fmt.Sprintf("%.0f", temp.asFloat(sess.Unit))
		conn.WriteJSON(cmd)
	})
}
