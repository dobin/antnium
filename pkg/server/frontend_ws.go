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
	a := FrontendWs{
		campaign,
		make(map[*websocket.Conn]bool),
		make(chan PacketInfo),
		websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}

	go a.Distributor() // FIXME good here?

	return a
}

// NewConnectionHandler is the entry point for new Frontend/UI websocket connections
func (a *FrontendWs) NewConnectionHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := a.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("FrontendWs: %s", err.Error())
		return
	}

	// WebSocket Authentication
	// first message should be the AdminApiKey
	var authToken AuthToken
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("FrontendWs read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Warnf("FrontendWs: could not decode auth: %v", message)
		return
	}
	if string(authToken) == a.campaign.AdminApiKey {
		a.registerWs(ws)
	} else {
		log.Warn("FrontendWs: incorrect key: " + authToken)
	}
}

// wsHandler is the entry point for new websocket connections
func (a *FrontendWs) registerWs(wsConn *websocket.Conn) {
	a.clients[wsConn] = true
}

// Distributor is a Thread which distributes data to all connected websocket clients. Lifetime: app
func (a *FrontendWs) Distributor() {
	for {
		packetInfo := <-a.channelDistributor
		websocketData := WebsocketData{
			packetInfo,
		}

		data, err := json.Marshal(websocketData)
		if err != nil {
			log.Error("Could not JSON marshal")
			continue
		}

		// send to every client that is currently connected
		for client := range a.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Debugf("FrontendWs::Distributor() error%s", err)
				log.Debugf("FrontendWs::Distributor() closing WS socket: %s", client.RemoteAddr())
				client.Close()
				delete(a.clients, client)
			}
		}
	}
}
