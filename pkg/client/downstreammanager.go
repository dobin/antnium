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

// startListeners will set up all downstreams which have a listening component as threads
func (dm *DownstreamManager) StartListeners(client *Client) {
	// Localtcp thread, new clients via downstreamLocaltcpChannel
	go dm.downstreamLocaltcp.startServer(dm.downstreamLocaltcpChannel)

	// If no client is given, dont send it to upstream. For unittesting.
	if client == nil {
		return
	}
	go func() {
		//downstreamList := <-dm.downstreamLocaltcpChannel
		<-dm.downstreamLocaltcpChannel       // Ignore new client for now
		client.SendDownstreams(dm.GetList()) // notify server of new downstream executors
	}()
}

func (dm *DownstreamManager) Do(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId == "client" {
		return dm.downstreamClient.do(packet)
	} else if strings.HasPrefix(packet.DownstreamId, "net") { // net#1
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
