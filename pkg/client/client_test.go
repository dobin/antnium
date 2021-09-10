package client

import (
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/server"
)

func makeSimpleCmdPacket(computerId string, packetId string, commandline string) model.Packet {
	arguments := make(model.PacketArgument)
	arguments["shelltype"] = "cmd"
	arguments["commandline"] = commandline
	response := make(model.PacketResponse)
	packet := model.NewPacket("exec", computerId, packetId, arguments, response)
	return packet
}

func TestClientExecWs(t *testing.T) {
	t.Parallel()

	port := "55191"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	packet := makeSimpleCmdPacket(computerId, "p42", "echo test")
	s.Middleware.AdminAddNewPacket(packet)

	// make server go
	go s.Serve()
	defer s.Shutdown()

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	client.Start()
	go client.Loop()

	// Wait for packet answer
	for {
		if s.Middleware.AdminGetAllPacket()[0].State == server.STATE_ANSWERED {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	// Check them
	packetInfos := s.Middleware.AdminGetAllPacket()
	if packetInfos[0].Packet.Arguments["commandline"] != "echo test" {
		t.Error("wrong packet")
	}
	if !strings.Contains(packetInfos[0].Packet.Response["stdout"], "test") {
		t.Error("wrong stdout")
	}

	if packetInfos[0].State != server.STATE_ANSWERED {
		t.Error("Wrong state")
	}
}

func TestClientParalellExecWs(t *testing.T) {
	t.Parallel()

	// Add two commands, one sleep, one echo
	port := "55192"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true
	defer s.Shutdown()

	// Make a example packet the client should receive
	packet := makeSimpleCmdPacket(computerId, "p42", "ping localhost")
	s.Middleware.AdminAddNewPacket(packet)
	packet = makeSimpleCmdPacket(computerId, "p43", "echo test")
	s.Middleware.AdminAddNewPacket(packet)

	// Start server
	go s.Serve()
	defer s.Shutdown()

	// Start client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ComputerId = computerId
	client.Start()
	go client.Loop()

	// Wait for packet answer
	for {
		if s.Middleware.AdminGetAllPacket()[1].State == server.STATE_ANSWERED {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	packetInfos := s.Middleware.AdminGetAllPacket()
	if len(packetInfos) != 2 {
		t.Error("Not 2")
		return
	}
	if packetInfos[1].Packet.Arguments["commandline"] != "echo test" {
		t.Error("wrong packet")
		return
	}
	if !strings.Contains(packetInfos[1].Packet.Response["stdout"], "test") {
		t.Error("wrong stdout")
		return
	}
	if packetInfos[1].State != server.STATE_ANSWERED {
		t.Error("Wrong state")
		return
	}
}