package client

import (
	"os"
	"strconv"
	"strings"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

// Only used to translate downstream specific downstreaminfos from children
// into a more managable struct.
type DownstreamInfo struct {
	Name string
	Info string
}

type DownstreamManager struct {
	downstreamClient     DownstreamClient
	downstreamClientInfo string

	downstreamLocaltcp        DownstreamLocaltcp
	downstreamLocaltcpChannel chan struct{} // Notify only
}

func MakeDownstreamManager() DownstreamManager {
	downstreamClient := MakeDownstreamClient()
	downstreamLocaltcp := MakeDownstreamLocaltcp("")

	downstreamManager := DownstreamManager{
		downstreamClient,
		"client.exe",

		downstreamLocaltcp,
		make(chan struct{}),
	}
	return downstreamManager
}

// startListeners will set up all downstreams which have a listening component as threads
func (dm *DownstreamManager) StartListeners(client *Client) {

	// Do it here for now, as it is always executed
	ex, err := os.Executable()
	if err != nil {
		log.Error("Error: " + err.Error())
	}
	pid := strconv.Itoa(os.Getpid())
	line := ex + ":" + pid + "\n"
	dm.downstreamClientInfo = line

	// Thread: new downstreams via downstreamLocaltcpChannel
	go dm.downstreamLocaltcp.startServer(dm.downstreamLocaltcpChannel)

	// Thread: receive new downstreams via local tcp, lifetime: app
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
		dm.downstreamClientInfo,
	}
	downstreamInfoTcp := dm.downstreamLocaltcp.DownstreamList()

	downstreams = append(downstreams, downstreamInfoClient)
	downstreams = append(downstreams, downstreamInfoTcp...)

	// Notify server
	client.SendDownstreams(downstreams) // notify server of new downstream executors
}
