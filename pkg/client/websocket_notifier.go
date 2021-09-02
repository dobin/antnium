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

type WebsocketNotifier struct {
	channel chan string

	state ClientState
	coder model.Coder

	config   *ClientConfig
	campaign *campaign.Campaign

	wsConn *websocket.Conn
}

func MakeWebsocketNotifier(config *ClientConfig, campaign *campaign.Campaign) WebsocketNotifier {
	coder := model.MakeCoder(campaign)

	u := WebsocketNotifier{
		make(chan string),
		MakeClientState(),
		coder,
		config,
		campaign,
		nil,
	}
	return u
}

func (d *WebsocketNotifier) Connect() error {
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

	// Thread: Receiver Thread
	go d.Start()

	return nil
}

func (d *WebsocketNotifier) IsConnected() bool {
	if d.wsConn == nil {
		return false
	} else {
		return true
	}
}

// Start is a Thread responsible for receiving notifications from server, lifetime:websocket connection
func (d *WebsocketNotifier) Start() {
	defer d.wsConn.Close()

	for {
		// Get notification (blocking)
		_, _, err := d.wsConn.ReadMessage()
		if err == nil {
			d.channel <- "notification"
		} else {
			d.channel <- "notification" // ALWAYS send back something, or upstream will get stuck
			log.Error("WS error, closed?")
			d.wsConn = nil
			break
		}
	}
}
