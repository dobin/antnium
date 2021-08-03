package client

import (
	"strings"

	"github.com/dobin/antnium/pkg/model"
)

type DownstreamInfo struct {
	Name string
	Info string
}

type DownstreamManager struct {
	downstreamClient DownstreamClient

	downstreamLocaltcp        DownstreamLocaltcp
	downstreamLocaltcpChannel chan struct{} // Notify only
}

func MakeDownstreamManager() DownstreamManager {
	downstreamClient := MakeDownstreamClient()
	downstreamLocaltcp := MakeDownstreamLocaltcp("")

	downstreamManager := DownstreamManager{
		downstreamClient,
		downstreamLocaltcp,
		make(chan struct{}),
	}
	return downstreamManager
}

// startListeners will set up all downstreams which have a listening component as threads
func (dm *DownstreamManager) StartListeners(client *Client) {
	// Thread: new downstreams via downstreamLocaltcpChannel
	go dm.downstreamLocaltcp.startServer(dm.downstreamLocaltcpChannel)

	// Thread: receive new downstreams via local tcp
	go func() {
		for {
			// Wait for newly announced TCP downstreams
			<-dm.downstreamLocaltcpChannel

			// Notify server
			dm.SendDownstreams(client)

			// TODO when to quit thread
		}
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

func (dm *DownstreamManager) SendDownstreams(client *Client) {
	downstreams := make([]DownstreamInfo, 0)
	downstreamInfoClient := DownstreamInfo{
		"client",
		"client.exe",
	}
	downstreamInfoTcp := dm.downstreamLocaltcp.DownstreamList()

	downstreams = append(downstreams, downstreamInfoClient)
	downstreams = append(downstreams, downstreamInfoTcp...)

	// Notify server
	client.SendDownstreams(downstreams) // notify server of new downstream executors
}
