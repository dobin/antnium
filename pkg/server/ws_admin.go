package server

/* Mostly based on
   https://rogerwelin.github.io/golang/websockets/gorilla/2018/03/13/golang-websockets.html
*/

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebsocketData is just a wrapper for PacketInfo atm
type WebsocketData struct {
	PacketInfo PacketInfo `json:"PacketInfo"`
}

type AuthToken string

type AdminWebSocket struct {
	clients     map[*websocket.Conn]bool
	adminapiKey string

	channel chan *WebsocketData
}

func MakeAdminWebSocket(adminApiKey string) AdminWebSocket {
	a := AdminWebSocket{
		make(map[*websocket.Conn]bool),
		adminApiKey,
		make(chan *WebsocketData),
	}

	go a.Distributor() // FIXME good here?

	return a
}

// wsHandler is the entry point for new websocket connections
func (a *AdminWebSocket) registerWs(wsConn *websocket.Conn) {
	a.clients[wsConn] = true
}

func (a *AdminWebSocket) broadcastPacket(packetInfo PacketInfo) {
	websocketData := WebsocketData{
		packetInfo,
	}
	a.channel <- &websocketData
}

// Distributor is a Thread which distributes data to all connected websocket clients. Lifetime: app
func (a *AdminWebSocket) Distributor() {
	for {
		guiData := <-a.channel

		data, err := json.Marshal(guiData)
		if err != nil {
			log.Error("Could not JSON marshal")
		}

		// send to every client that is currently connected
		for client := range a.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("AdminWebsocket error: %s", err)
				client.Close()
				delete(a.clients, client)
			}
		}
	}
}
