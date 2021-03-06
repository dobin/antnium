package client

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/server"
)

func makeSimpleCmdPacket(clientId string, packetId string, commandline string) model.Packet {
	arguments := make(model.PacketArgument)

	if runtime.GOOS == "linux" {
		arguments["shelltype"] = "bash"
	} else if runtime.GOOS == "windows" {
		arguments["shelltype"] = "cmd"
	} else if runtime.GOOS == "darwin" {
		arguments["shelltype"] = "zsh"
	}
	arguments["commandline"] = commandline
	response := make(model.PacketResponse)
	packet := model.NewPacket("exec", clientId, packetId, arguments, response)
	return packet
}

// TestClientExecWs tests if the client executes a packet from the perspective of a server
func TestClientExecWs(t *testing.T) {
	//t.Parallel()

	port, _ := common.FreePort()
	clientId := "clientid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true

	// Make a example packet the client should receive
	packet := makeSimpleCmdPacket(clientId, "p42", "echo test")
	s.Middleware.FrontendAddNewPacket(&packet, "")

	// make server go
	go s.Serve()
	defer s.Shutdown()

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ClientId = clientId
	client.Start()
	go client.Loop()

	// Wait for packet answer
	for {
		if s.Middleware.FrontendAllPacket()[0].State == server.STATE_ANSWERED {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	// Check them
	packetInfos := s.Middleware.FrontendAllPacket()
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

// TestClientParalellExecWs starts two execs in a client and checks that the quick one returns first
func TestClientParalellExecWs(t *testing.T) {
	//t.Parallel()

	// Add two commands, one sleep, one echo
	port, _ := common.FreePort()
	clientId := "clientid-23"

	// Server in background, checking via client
	s := server.NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true
	packetA := makeSimpleCmdPacket(clientId, "p42", "ping localhost")
	s.Middleware.FrontendAddNewPacket(&packetA, "")
	packetB := makeSimpleCmdPacket(clientId, "p43", "echo test")
	s.Middleware.FrontendAddNewPacket(&packetB, "")
	defer s.Shutdown()
	go s.Serve()

	// Start client
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ClientId = clientId
	client.Start()
	go client.Loop()

	// Wait for packet answer
	n := 0
	for {
		n += 1
		if s.Middleware.FrontendAllPacket()[1].State == server.STATE_ANSWERED {
			break
		}
		if n == 10 {
			t.Errorf("Packet not answerd in time: %v", s.Middleware.FrontendAllPacket())
			return
		}
		time.Sleep(time.Millisecond * 50)
	}

	packetInfos := s.Middleware.FrontendAllPacket()
	if len(packetInfos) != 3 { // two packets and a clientinfo
		t.Error("Not 3")
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

	// test if we received only the quick to respond answer
	if packetInfos[0].State == server.STATE_ANSWERED {
		t.Error("Wrong state")
		return
	}
}
