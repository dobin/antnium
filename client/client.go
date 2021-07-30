package client

import (
	"errors"

	"github.com/dobin/antnium/model"
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

	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-s.upstream.channel

		// Select appropriate downstream channel
		c := s.downstreamManager.GetFor(p)
		// Send it to the downstream
		c <- p
		// Receive answer
		p = <-c

		// Send answer back to server
		s.upstream.channel <- p
	}
}
