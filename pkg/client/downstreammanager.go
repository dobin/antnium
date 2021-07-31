package client

import (
	"github.com/dobin/antnium/pkg/model"
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
	go dm.downstreamLocaltcp.startServer()
}

func (dm *DownstreamManager) do(packet model.Packet) (model.Packet, error) {
	channelId, ok := packet.Arguments["channelId"]
	if !ok {
		return dm.downstreamClient.do(packet)
	}

	if channelId == "client" {
		return dm.downstreamClient.do(packet)
	} else if channelId == "net#1" {
		return dm.downstreamLocaltcp.do(packet)
	} else {
		return dm.downstreamClient.do(packet)
	}
}

func (dm *DownstreamManager) GetList() []string {
	ret := make([]string, 0)

	ret = append(ret, "client")
	ret = append(ret, "net#1")

	return ret
}
