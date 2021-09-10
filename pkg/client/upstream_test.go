package client

import (
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/server"
)

func TestUpstreamServerHttp(t *testing.T) {
	port := "55041"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)

	// disable websocket for HTTP only
	s.Campaign.ClientUseWebsocket = false

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	// make server go
	go s.Serve()

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false
	client.Config.ComputerId = computerId
	client.Start()

	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}

	s.Shutdown()
}

func TestUpstreamServerWs(t *testing.T) {
	port := "55141"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)

	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	// make server go
	go s.Serve()

	// make client
	client := NewClient()
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

func TestUpstreamServerWsConnect(t *testing.T) {
	port := "55046"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Test: No server connection
	// TODO

	time.Sleep(10 * time.Millisecond)
	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	// Make a example packet the client should receive
	s.Campaign.ClientUseWebsocket = true
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// Test: Client connected
	// expect packet to be received upon connection (its already added)
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet
}

func TestUpstreamServerWsReconnect(t *testing.T) {
	port := "55047"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	// Make a example packet the client should receive
	s.Campaign.ClientUseWebsocket = true
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Get packet
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet

	if !client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client not connected?")
		return
	}

	// Kill server
	time.Sleep(time.Millisecond * 10) // give client time to answer
	s.Shutdown()
	time.Sleep(time.Millisecond * 10) // give time to really shutdown

	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// 2nd server
	s = server.NewServer("127.0.0.1:" + port)
	// Make a example packet the client should receive
	s.Campaign.ClientUseWebsocket = true
	arguments = make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response = make(model.PacketResponse)
	packet = model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo = server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// Test: Client reconnected
	// expect packet to be received upon connection (its already added)
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet

	if !client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client not connected?")
		return
	}
}

func TestUpstreamServerHttpConnect(t *testing.T) {
	port := "55077"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Test: No server connection
	// TODO

	time.Sleep(10 * time.Millisecond)
	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	// Make a example packet the client should receive
	s.Campaign.ClientUseWebsocket = true
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// Test: Client connected
	// expect packet to be received upon connection (its already added)
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
	}
	client.UpstreamManager.Channel <- packet
}
