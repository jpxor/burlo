package main

import (
	"burlo/pkg/models/controller"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func pushThermostatToDashboards(tstat controller.Thermostat) {
	ws.write(tstat)
}

func pushWeatherToDashboards(weather Weather) {
	ws.write(weather)
}

func pushSetpointToDashboards(value float32) {
	var command = struct {
		Command string `json:"command"`
		Id      string `json:"id"`
		Html    string `json:"html"`
	}{
		Command: "setInnerHTML",
		Id:      "setpoint-value",
		Html:    fmt.Sprintf("%.0f", value),
	}
	ws.write(command)
}

var ws = WebSocketConnections{
	connections: make(map[*websocket.Conn]bool),
}

type WebSocketConnections struct {
	connections map[*websocket.Conn]bool
	mutex       sync.Mutex
}

func (ws *WebSocketConnections) newConnection(conn *websocket.Conn) {
	conn.SetCloseHandler(func(code int, text string) error {
		ws.mutex.Lock()
		defer ws.mutex.Unlock()
		delete(ws.connections, conn)
		return nil
	})
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.connections[conn] = true
}

func (ws *WebSocketConnections) write(jsonObj interface{}) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	for conn, ok := range ws.connections {
		if ok {
			err := conn.WriteJSON(jsonObj)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func AcceptWebsocket() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		ws.newConnection(conn)
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(messageType, string(p))
		}
	}
}
