package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/dobin/antnium/client"
	"github.com/dobin/antnium/model"
)

func TestServerClientIntegration(t *testing.T) {
	port := "55001"
	packetId := "packetid-42"
	computerId := "computerid-23"
	s := NewServer("127.0.0.1:" + port)

	// Make a example command the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	command := model.NewCommand("test", computerId, packetId, arguments, response)
	packetInfo := NewPacketInfo(command, STATE_RECORDED)
	s.packetDb.add(packetInfo)

	// make server go
	go s.Serve()

	// create client, receive the command we added above
	// This tests most of the stuff (encryption, encoding, campaign data, server paths and more)
	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Config.ComputerId = computerId
	command, err := c.GetCommand()
	if err != nil {
		t.Errorf("Error when receiving command: " + err.Error())
	}
	if command.PacketId != packetId {
		t.Errorf("Command received, but wrong packetid: %s", command.PacketId)
	}
	if command.Arguments["arg0"] != "value0" {
		t.Errorf("Command received, but wrong args: %v", command.Arguments)
	}
}

func TestServerAuthAdmin(t *testing.T) {
	var err error

	// Start server in the background
	port := "55002"
	s := NewServer("127.0.0.1:" + port)
	go s.Serve()

	// Create a default (non authenticated) HTTP client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Test Admin
	r, _ := http.NewRequest("GET", "http://127.0.0.1:55002/admin/commands", nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 200 {
		t.Errorf("Could access admin API without authentication")
	}
}

func TestServerAuthClient(t *testing.T) {
	var err error
	var url string
	packetId := "packetid-42"
	computerId := "computerid-23"

	// Start server in the background
	port := "55000"
	s := NewServer("127.0.0.1:" + port)

	// Make a example command the client should receive
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	command := model.NewCommand("test", computerId, packetId, arguments, response)
	packetInfo := NewPacketInfo(command, STATE_RECORDED)
	s.packetDb.add(packetInfo)

	go s.Serve()

	// Create a default (non authenticated) HTTP client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:" + port
	c.Config.ComputerId = computerId

	// Test Client: No key
	url = c.CommandGetUrl()
	r, _ := http.NewRequest("GET", url, nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		t.Errorf("Error accessing server api with url: " + url)
	}
	if resp.StatusCode == 200 {
		t.Errorf("Could access server API though i should not: " + url)
	}

	// Test Client: Correct key
	command, err = c.GetCommand()
	if err != nil {
		t.Errorf("Could not get command: " + err.Error())
	}
	if command.Command != "test" {
		t.Errorf("Recv command err")
	}
	if command.ComputerId != computerId {
		t.Errorf("Recv command err")
	}
	if command.PacketId != packetId {
		t.Errorf("Recv command err")
	}

	// Test: Static
	/*
		url = c.CommandGetUrl()
		r, _ = http.NewRequest("GET", url, nil)
		resp, err = unauthHttp.Do(r)
		if err != nil {
			t.Errorf("Error accessing static with url: " + url)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Could access static: " + url)
		}
	*/

	// Test: Upload?
}
