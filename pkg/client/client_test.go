package client

import (
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/server"
	log "github.com/sirupsen/logrus"
)

func TestClientExecWs(t *testing.T) {
	// Add two commands, one sleep, one echo

	port := "55191"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)

	// disable websocket for HTTP only
	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["commandline"] = "echo test"
	arguments["shelltype"] = "cmd"
	response := make(model.PacketResponse)
	packet := model.NewPacket("exec", computerId, "p42", arguments, response)
	//packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	//s.Middleware.AddPacketInfo(packetInfo)
	s.Middleware.AdminAddNewPacket(packet)

	// make server go
	go s.Serve()

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

	s.Shutdown()
}

func TestClientParalellExecWs(t *testing.T) {
	// Add two commands, one sleep, one echo
	port := "55192"
	computerId := "computerid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)

	// disable websocket for HTTP only
	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["commandline"] = "ping localhost"
	arguments["shelltype"] = "cmd"
	response := make(model.PacketResponse)
	packet := model.NewPacket("exec", computerId, "p42", arguments, response)
	//packetInfo := server.NewPacketInfo(packet, server.STATE_RECORDED)
	//s.Middleware.AddPacketInfo(packetInfo)
	s.Middleware.AdminAddNewPacket(packet)

	arguments = make(model.PacketArgument)
	arguments["commandline"] = "echo test"
	arguments["shelltype"] = "cmd"
	response = make(model.PacketResponse)
	packet = model.NewPacket("exec", computerId, "p43", arguments, response)
	//packetInfo = server.NewPacketInfo(packet, server.STATE_RECORDED)
	//s.Middleware.AddPacketInfo(packetInfo)
	s.Middleware.AdminAddNewPacket(packet)

	// make server go
	go s.Serve()

	// Test Localtcp Downstream
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
	log.Infof("%v", packetInfos)
	if packetInfos[1].Packet.Arguments["commandline"] != "echo test" {
		t.Error("wrong packet")
	}
	if !strings.Contains(packetInfos[1].Packet.Response["stdout"], "test") {
		t.Error("wrong stdout")
	}

	if packetInfos[1].State != server.STATE_ANSWERED {
		t.Error("Wrong state")
	}

	s.Shutdown()
}
