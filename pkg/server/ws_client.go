package server

import (
	"net/http"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ClientWebSocket struct {
	clients    map[string]*websocket.Conn // ComputerId:WebsocketConnection
	wsUpgrader websocket.Upgrader
	coder      model.Coder
	campaign   *campaign.Campaign
}

func MakeClientWebSocket(campaign *campaign.Campaign) ClientWebSocket {
	a := ClientWebSocket{
		clients:    make(map[string]*websocket.Conn),
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		coder:      model.MakeCoder(campaign),
		campaign:   campaign,
	}
	return a
}

// wsHandler is the entry point for new websocket connections
func (a *ClientWebSocket) registerWs(computerId string, ws *websocket.Conn) {
	// register client as auth succeeded
	a.clients[computerId] = ws

	// Thread which reads from the connection, to:
	// * Fulfill Websocket requirement
	// * Detect closed websocket connections
	// See https://pkg.go.dev/github.com/gorilla/websocket?utm_source=godoc#hdr-Control_Messages
	// Lifetime: Websocket connection
	go func() {
		for {
			/*_, packetReader, err := ws.NextReader()
			if err != nil {
				ws.Close()
				a.clients[authToken.ComputerId] = nil
				break
			}*/

			//packetData, err := packetReader.Read()

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

			log.Info("AAA 2: %v", packet)

			//a.server.AddNewClientPacket(packet)
		}
	}()
}

func (a *ClientWebSocket) TryNotify(packet *model.Packet) bool {
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

	/*err := clientConn.WriteMessage(websocket.TextMessage, []byte("notification"))
	if err != nil {
		log.Infof("Websocket for host %s closed when trying to write: %s", packet.ComputerId, err.Error())
		return
	}*/
	log.Infof("Client %s notified about new packet via WS", packet.ComputerId)

	return true
}
