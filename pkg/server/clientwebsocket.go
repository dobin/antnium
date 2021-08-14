package server

import (
	"encoding/json"
	"net/http"

	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ClientWebSocket struct {
	clients    map[string]*websocket.Conn
	wsUpgrader websocket.Upgrader
}

func MakeClientWebSocket() ClientWebSocket {
	a := ClientWebSocket{
		make(map[string]*websocket.Conn),
		websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return a
}

// wsHandler is the entry point for new websocket connections
func (a *ClientWebSocket) wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("New WS connection")

	ws, err := a.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Websocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	var authToken model.ClientWebSocketAuth
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("Websocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Errorf("WebSocket: could not decode auth: %v", message)
		return
	}

	if authToken.Key != "antnium" {
		log.Warn("WebSocket: incorrect key: " + authToken.Key)
		return
	}
	// register client as auth succeeded
	a.clients[authToken.ComputerId] = ws
}

func (a *ClientWebSocket) TryNotify(packet *model.Packet) {
	client, ok := a.clients[packet.ComputerId]
	if !ok {
		// All ok, not connected to ws
		return
	}
	client.WriteMessage(websocket.TextMessage, []byte("notification"))
	log.Infof("Notified: %s about new packet", packet.ComputerId)
}
