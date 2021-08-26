package server

import (
	"encoding/json"
	"net/http"

	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ClientWebSocket struct {
	clients    map[string]*websocket.Conn // ComputerId:WebsocketConnection
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
	ws, err := a.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("ClientWebsocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	var authToken model.ClientWebSocketAuth
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("ClientWebsocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Errorf("ClientWebsocket: could not decode auth: %v", message)
		return
	}
	if authToken.Key != "antnium" {
		log.Warn("ClientWebsocket: incorrect key: " + authToken.Key)
		return
	}

	// register client as auth succeeded
	a.clients[authToken.ComputerId] = ws

	// Thread which reads from the connection, to:
	// * Fulfill Websocket requirement
	// * Detect closed websocket connections
	// See https://pkg.go.dev/github.com/gorilla/websocket?utm_source=godoc#hdr-Control_Messages
	// Lifetime: Websocket connection
	go func() {
		for {
			if _, _, err := ws.NextReader(); err != nil {
				ws.Close()
				a.clients[authToken.ComputerId] = nil
				break
			}
		}
	}()
}

func (a *ClientWebSocket) TryNotify(packet *model.Packet) {
	clientConn, ok := a.clients[packet.ComputerId]
	if !ok {
		// All ok, not connected to ws
		return
	}
	err := clientConn.WriteMessage(websocket.TextMessage, []byte("notification"))
	if err != nil {
		log.Infof("Websocket for host %s closed when trying to write: %s", packet.ComputerId, err.Error())
	}
	log.Infof("Notified: %s about new packet", packet.ComputerId)
}
