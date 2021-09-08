package server

import (
	"testing"

	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/model"
)

func TestConnectorHttp(t *testing.T) {
	port := "55044"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := NewServer("127.0.0.1:" + port)

	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	// make server go
	go s.Serve()

	// make client
	client := client.NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	client.Start()

	// expect packet to be received upon connection (its already added)
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet

	// Add a test packet via Admin REST
	s.Middleware.AdminAddNewPacket(packet)

	// Expect it
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet
}
