package server

import (
	"testing"

	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/common"
)

func TestConnectorHttp(t *testing.T) {
	//defer goleak.VerifyNone(t)

	port, _ := common.FreePort()
	clientId := "clientid-23"

	// Server in background, checking via client
	s := NewServer("127.0.0.1:" + port)
	defer s.Shutdown()
	s.Campaign.ClientUseWebsocket = false // Test: REST
	packetA := makeSimpleTestPacket(clientId, "001")
	s.Middleware.FrontendAddNewPacket(packetA, "")
	go s.Serve()

	// make client
	client := client.NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false // Test: REST
	client.Config.ClientId = clientId
	client.Start()
	defer client.Shutdown()

	// expect packet to be received upon connection (its already added)
	packetB := <-client.UpstreamManager.ChannelIncoming
	if packetB.PacketId != "001" || packetB.ClientId != clientId {
		t.Error("Err")
		return
	}

	// Add a test packet via Frontend REST
	packetC := makeSimpleTestPacket(clientId, "002")
	s.Middleware.FrontendAddNewPacket(packetC, "")

	// Expect it
	packetD := <-client.UpstreamManager.ChannelIncoming
	if packetD.PacketId != "002" || packetD.ClientId != clientId {
		t.Error("Err")
		return
	}
}

func TestConnectorWs(t *testing.T) {
	//defer goleak.VerifyNone(t)

	port, _ := common.FreePort()
	clientId := "clientid-23"

	// Server in background, checking via client
	s := NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true
	packetA := makeSimpleTestPacket(clientId, "001")
	s.Middleware.FrontendAddNewPacket(packetA, "")
	defer s.Shutdown()
	go s.Serve()

	// make client
	client := client.NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = true
	client.Config.ClientId = clientId
	client.Start()
	defer client.Shutdown()

	// expect packet to be received upon connection (its already added)
	packetB := <-client.UpstreamManager.ChannelIncoming
	if packetB.PacketId != "001" || packetB.ClientId != clientId {
		t.Error("Err")
	}

	// Add a test packet via Frontend REST
	packetC := makeSimpleTestPacket(clientId, "002")
	s.Middleware.FrontendAddNewPacket(packetC, "")

	// Expect it
	packetD := <-client.UpstreamManager.ChannelIncoming
	if packetD.PacketId != "002" || packetD.ClientId != clientId {
		t.Error("Err")
	}

}
