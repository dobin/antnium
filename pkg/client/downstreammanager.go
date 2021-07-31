package client

import (
	"strings"

	"github.com/dobin/antnium/pkg/model"
)

type DownstreamInfo struct {
	Name    string
	Details string
}

type DownstreamManager struct {
	downstreamClient DownstreamClient

	downstreamLocaltcp        DownstreamLocaltcp
	downstreamLocaltcpChannel chan []string
}

func MakeDownstreamManager() DownstreamManager {
	downstreamClient := MakeDownstreamClient()
	downstreamLocaltcp := MakeDownstreamLocaltcp()

	downstreamManager := DownstreamManager{
		downstreamClient,
		downstreamLocaltcp,
		make(chan []string),
	}
	return downstreamManager
}

func (dm *DownstreamManager) start(client *Client) {
	go dm.downstreamLocaltcp.startServer(dm.downstreamLocaltcpChannel)
	go func() {
		//downstreamList := <-dm.downstreamLocaltcpChannel
		<-dm.downstreamLocaltcpChannel // Ignore it for now
		client.SendDownstreams(dm.GetList())
	}()
}

func (dm *DownstreamManager) do(packet model.Packet) (model.Packet, error) {
	channelId, ok := packet.Arguments["channelId"]
	if !ok {
		return dm.downstreamClient.do(packet)
	}

	if channelId == "client" {
		return dm.downstreamClient.do(packet)
	} else if strings.HasPrefix(channelId, "net") { // net#1
		return dm.downstreamLocaltcp.do(packet)
	} else {
		return dm.downstreamClient.do(packet)
	}
}

func (dm *DownstreamManager) GetList() []string {
	ret := make([]string, 0)

	ret = append(ret, "client")
	ret = append(ret, dm.downstreamLocaltcp.getList()...)

	return ret
}
