package server

/* Mostly based on
   https://rogerwelin.github.io/golang/websockets/gorilla/2018/03/13/golang-websockets.html
*/

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type GuiData struct {
	Reason     string `json:"Reason"`
	ComputerId string `json:"ComputerId"`
}

type AdminWebSocket struct {
	clients     map[*websocket.Conn]bool
	adminapiKey string
}

func MakeAdminWebSocket(adminApiKey string) AdminWebSocket {
	a := AdminWebSocket{
		make(map[*websocket.Conn]bool),
		adminApiKey,
	}
	return a
}

/****/

var broadcast = make(chan *GuiData)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/****/

type AuthToken string

func (a *AdminWebSocket) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// WebSocket Authentication
	// first message should be the AdminApiKey
	var authToken AuthToken
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("Websocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Warn("WebSocket: could not decode auth")
		return
	}
	if string(authToken) == a.adminapiKey {
		// register client as auth succeeded
		a.clients[ws] = true
	} else {
		log.Warn("WebSocket: incorrect key: " + authToken)
	}
}

func (a *AdminWebSocket) broadcastCmd(reason string, computerId string) {
	guiData := GuiData{
		reason,
		computerId,
	}
	broadcast <- &guiData
	fmt.Printf("Sending to WS: %v\n", guiData)
}

func (a *AdminWebSocket) Distributor() {
	for {
		guiData := <-broadcast

		data, err := json.Marshal(guiData)
		if err != nil {
			log.Error("Could not JSON marshal")
		}

		// send to every client that is currently connected
		for client := range a.clients {
			err := client.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(a.clients, client)
			}
		}
	}
}
