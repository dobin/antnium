package client

import (
	"github.com/dobin/antnium/model"
)

type DownstreamManager struct {
	downstreamClient   DownstreamClient
	downstreamLocaltcp DownstreamLocaltcp
}

func MakeDownstreamManager() DownstreamManager {
	downstreamClient := MakeDownstreamClient()
	downstreamLocaltcp := MakeDownstreamLocaltcp()

	downstreamManager := DownstreamManager{
		downstreamClient,
		downstreamLocaltcp,
	}
	return downstreamManager
}

func (dm *DownstreamManager) start() {
	go dm.downstreamClient.start()
	go dm.downstreamLocaltcp.start()
}

func (dm *DownstreamManager) GetFor(p model.Packet) chan model.Packet {
	return dm.downstreamClient.channel
}
