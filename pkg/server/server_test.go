package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
)

func makeSimpleTestPacket(clientId string, packetId string) *model.Packet {
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", clientId, packetId, arguments, response)
	return &packet
}

// TestServerClientIntegrationRest will check if client and server can communicate via HTTP.
func TestServerClientIntegrationRest(t *testing.T) {
	port, _ := common.FreePort()
	clientId := "clientid-23"
	packetId := "packetid-42"

	s := NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = false // Test: REST
	packet := makeSimpleTestPacket(clientId, packetId)
	s.Middleware.FrontendAddNewPacket(packet, "")
	go s.Serve()

	// create client, receive the packet we added above
	// This tests most of the stuff (encryption, encoding, campaign data, server paths and more)
	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Campaign.ClientUseWebsocket = false // Test: REST
	c.Config.ClientId = clientId
	c.Start()

	answerPacket := <-c.UpstreamManager.ChannelIncoming
	if answerPacket.PacketId != packetId {
		t.Errorf("Packet received, but wrong packetid: %s", packet.PacketId)
		return
	}
	if answerPacket.Arguments["arg0"] != "value0" {
		t.Errorf("Packet received, but wrong args: %v", packet.Arguments)
		return
	}
	s.Shutdown()
}

// TestServerClientIntegrationWebsocket will check if client and server can communicate via websocket
func TestServerClientIntegrationWebsocket(t *testing.T) {
	port, _ := common.FreePort()
	clientId := "clientid-23"
	packetId := "packetid-42"

	// Server
	s := NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true // Test: WS
	go s.Serve()

	// Let the client connect
	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Campaign.ClientUseWebsocket = true // Test: WS
	c.Campaign.DoClientInfo = false      // Theres some kind of race condition going on
	c.Config.ClientId = clientId
	c.Start()

	// Send test packet via admin interface
	packet := makeSimpleTestPacket(clientId, packetId)
	json_data, err := json.Marshal(packet)
	if err != nil {
		t.Errorf("Error when receiving packet: " + err.Error())
		return
	}
	url := c.Campaign.ServerUrl + "/admin/addPacket/username"
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		t.Errorf("Unittest error: %s", err.Error())
		return
	}
	req.Header.Set("Authorization", s.config.AdminApiKey)
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("Unittest error: %s", err.Error())
		return
	}
	if res.StatusCode != 200 {
		t.Errorf("Unittest error: HTTP response not status 200: %v", res)
		return
	}

	// this should return immediately, as notified via websocket
	var p model.Packet
	p = <-c.UpstreamManager.ChannelIncoming

	if p.PacketId != packetId {
		t.Errorf("Unittest error: Packet received, but wrong packetid: %s", packet.PacketId)
		return
	}
	if p.Arguments["arg0"] != "value0" {
		t.Errorf("Unittest error: Packet received, but wrong args: %v", packet.Arguments)
		return
	}

	s.Shutdown()
}

func TestServerAuthAdmin(t *testing.T) {
	var err error

	// Start server in the background
	port, _ := common.FreePort()
	s := NewServer("127.0.0.1:" + port)
	go s.Serve()

	time.Sleep(100 * time.Millisecond)
	// Create a default (non authenticated) REST client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Test Admin
	r, _ := http.NewRequest("GET", "http://127.0.0.1:"+port+"/admin/packets", nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		t.Errorf("Unittest error: HTTP get error: " + err.Error())
		return
	}
	if resp.StatusCode == 200 {
		t.Errorf("Unittest error: Could access admin API without authentication")
		return
	}

	s.Shutdown()
}

/*
func TestServerAuthClient(t *testing.T) {
	var err error
	packetId := "packetid-42"
	clientId := "clientid-23"

	// Start server in the background
	port := "55003"
	s := NewServer("127.0.0.1:" + port)

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", clientId, packetId, arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	s.middleware.packetDb.add(packetInfo)

	go s.Serve()

	c := client.NewClient()
	c.Campaign.ProxyUrl = "" // Always disable proxy
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Config.ClientId = clientId
	c.Start()

	// Try first with invalid key (consumes one packet)
	origEncKey := c.Campaign.EncKey
	c.Campaign.EncKey = []byte("12345678123456781234567812345678")
	//packet, err = c.UpstreamManager.UpstreamRest.GetPacket()
	packet = <-c.UpstreamManager.ChannelIncoming
	c.UpstreamManager.ChannelIncoming <- packet

	if err == nil {
		t.Errorf("Unittest error: Could get packet with wrong enckey! %v", packet)
		return
	}
	c.Campaign.EncKey = origEncKey

	// Test Client: Correct key (consumes one packet)
	//s.middleware.packetDb.add(packetInfo) // changing var of a thread, dangerous but works
	//packet, err = c.UpstreamManager.UpstreamRest.GetPacket()
	packet = <-c.UpstreamManager.ChannelIncoming
	c.UpstreamManager.ChannelIncoming <- packet

	if err != nil {
		t.Errorf("Unittest error: Could not get packet: " + err.Error())
		return
	}
	if packet.PacketType != "test" {
		t.Errorf("Unittest error: Recv packet err")
		return
	}
	if packet.ClientId != clientId {
		t.Errorf("Unittest error: Recv packet err")
		return
	}
	if packet.PacketId != packetId {
		t.Errorf("Unittest error: Recv packet err")
		return
	}

	// Test Client: Wrong key
	origApiKey := c.Campaign.ApiKey
	c.Campaign.ApiKey = "not42"
	//packet, err = c.UpstreamManager.UpstreamRest.GetPacket()
	packet = <-c.UpstreamManager.ChannelIncoming
	c.UpstreamManager.ChannelIncoming <- packet

	if err == nil {
		t.Errorf("Unittest error: Could get packet with wrong apikey!")
		return
	}
	c.Campaign.ApiKey = origApiKey

	// Test: Upload?
	// Test: Static?
}
*/
