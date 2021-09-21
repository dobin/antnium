package server

/* Mostly based on
   https://rogerwelin.github.io/golang/websockets/gorilla/2018/03/13/golang-websockets.html
*/

import (
	"encoding/json"
	"net/http"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebsocketData is just a wrapper for PacketInfo atm
type WebsocketData struct {
	PacketInfo PacketInfo `json:"PacketInfo"`
}

type AuthToken string

type FrontendWs struct {
	campaign           *campaign.Campaign
	clients            map[*websocket.Conn]bool
	channelDistributor chan PacketInfo

	wsUpgrader websocket.Upgrader
}

func MakeFrontendWs(campaign *campaign.Campaign) FrontendWs {
	f := FrontendWs{
		campaign,
		make(map[*websocket.Conn]bool),
		make(chan PacketInfo),
		websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}

	go f.Distributor() // FIXME good here?

	return f
}

// NewConnectionHandler is the entry point for new Frontend/UI websocket connections
func (f *FrontendWs) NewConnectionHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := f.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("FrontendWs: could not upgrade HTTP to websocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	// first message should be the AdminApiKey
	var authToken AuthToken
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("FrontendWs: read initial authentication error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Warnf("FrontendWs: could not decode authentication: %v", message)
		return
	}
	if string(authToken) != f.campaign.AdminApiKey {
		log.Warn("FrontendWs: incorrect key for authentication: " + authToken)
		return
	}

	f.registerWs(ws)
}

// wsHandler is the entry point for new websocket connections
func (f *FrontendWs) registerWs(wsConn *websocket.Conn) {
	f.clients[wsConn] = true
}

// Distributor is a Thread which distributes data to all connected websocket clients. Lifetime: app
func (f *FrontendWs) Distributor() {
	for {
		packetInfo, ok := <-f.channelDistributor
		if !ok {
			break
		}
		websocketData := WebsocketData{
			packetInfo,
		}

		data, err := json.Marshal(websocketData)
		if err != nil {
			log.Error("FrontendWs: Could not JSON marshal")
			continue
		}

		// send to every client that is currently connected
		for client := range f.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Debugf("FrontendWs::Distributor() error: %s", err)
				log.Debugf("FrontendWs::Distributor() closing WS socket: %s", client.RemoteAddr())
				client.Close()
				delete(f.clients, client)
			}
		}
	}
}

func (f *FrontendWs) Shutdown() {
	close(f.channelDistributor)
}
