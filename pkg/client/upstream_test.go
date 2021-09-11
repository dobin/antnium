package client

import (
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/server"
)

func makeSimpleTestPacket(computerId string, packetId string) *server.PacketInfo {
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", computerId, packetId, arguments, response)
	packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	return &packetInfo
}

// Test Upstream REST connection with running server
func TestUpstreamServerRest(t *testing.T) {
	t.Parallel()

	port := "55041"
	packetId := "packetid-421"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = false // Test: REST
	packetInfo := makeSimpleTestPacket(computerId, packetId)
	s.Middleware.AddPacketInfo(packetInfo)
	defer s.Shutdown()
	go s.Serve()

	// Test: Upstream Rest
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false // Test: REST
	client.Config.ComputerId = computerId
	client.Start()

	// Test: Just receive, no execute
	packet := <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}
}

// Test Upstream websocket connection with running server
func TestUpstreamServerWs(t *testing.T) {
	t.Parallel()

	port := "55141"
	packetId := "packetid-422"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true // Test: Websocket
	packetInfo := makeSimpleTestPacket(computerId, packetId)
	s.Middleware.AddPacketInfo(packetInfo)
	defer s.Shutdown()
	go s.Serve()

	// Test: Upstream Ws
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true // Test: Websocket
	client.Config.ComputerId = computerId
	client.Start()

	// Test: expect packet to be received upon connection (its already added)
	packet := <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}

	// Add a test packet via Admin REST
	s.Middleware.AdminAddNewPacket(packet)

	// Test: Expect it
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}
}

// Test Upstream REST connection with server first down and started later
func TestUpstreamServerWsConnectLate(t *testing.T) {
	t.Parallel()

	port := "55046"
	packetId := "packetid-423"
	computerId := "computerid-23"

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true // Test: WS
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Test: should have no server connection
	time.Sleep(10 * time.Millisecond)
	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	defer s.Shutdown()
	s.Campaign.ClientUseWebsocket = true // Test: WS
	packetInfo := makeSimpleTestPacket(computerId, packetId)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// Test: Client connected
	packet := <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}
}

// Test Upstream websocket connection by killing the server once
func TestUpstreamServerWsReconnect(t *testing.T) {
	t.Parallel()

	port := "55047"
	packetId1 := "packetid-42a"
	packetId2 := "packetid-42b"
	computerId := "computerid-23"

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true // Test: WS
	packetInfo := makeSimpleTestPacket(computerId, packetId1)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Get packet
	packet := <-client.UpstreamManager.Channel
	if packet.PacketId != packetId1 || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}

	// Should be true anyway, as we waited for a packet above
	if !client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client not connected?")
		return
	}

	// Kill server
	time.Sleep(time.Millisecond * 10) // give client time to answer
	s.Shutdown()
	time.Sleep(time.Millisecond * 10) // give time to really shutdown

	// Check if we lost connection
	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// Start 2nd server
	s = server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true
	packetInfo = makeSimpleTestPacket(computerId, packetId2) // make sure to take another packetId here
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()

	// Test: Client reconnected
	// expect packet to be received upon connection (its already added)
	packet = <-client.UpstreamManager.Channel
	if packet.PacketId != packetId2 || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}

	// Should be true, but test it anyway
	if !client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client not connected?")
		return
	}

	s.Shutdown()
}

// Test Upstream REST with server first down and started later
func TestUpstreamServerRestConnectLate(t *testing.T) {
	t.Parallel()

	port := "55077"
	packetId := "packetid-42"
	computerId := "computerid-23"

	// make client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false // Test: REST
	client.Config.ComputerId = computerId
	go client.Start() // start in background, as it tries to connect

	// Test: should have no server connection
	time.Sleep(10 * time.Millisecond)
	if client.UpstreamManager.UpstreamWs.Connected() {
		t.Error("Client connected?")
		return
	}

	// Start Server
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = false // Test: REST
	packetInfo := makeSimpleTestPacket(computerId, packetId)
	s.Middleware.AddPacketInfo(packetInfo)
	go s.Serve()
	defer s.Shutdown()

	// Test: Client connected
	// expect packet to be received upon connection (its already added)
	packet := <-client.UpstreamManager.Channel
	if packet.PacketId != packetId || packet.ComputerId != computerId {
		t.Error("Err")
		return
	}
}
