package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketConnections struct {
	connections map[*websocket.Conn]Session
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
	ws.connections[conn] = Session{
		Unit: C,
	}
}

func (ws *WebSocketConnections) forEach(fn func(*websocket.Conn, Session)) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	for conn, sess := range ws.connections {
		fn(conn, sess)
	}
	return nil
}
