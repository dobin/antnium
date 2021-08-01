package client

import (
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
)

func TestDownstreamClient(t *testing.T) {
	// Test default Downstream: "Client"
	client := NewClient()
	packet := makeTestPacket()
	packet, err := client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet")
	}
	if !strings.Contains(packet.Response["stdout"], "unreal") {
		t.Errorf("Incorrect output")
	}
}

func TestDownstreamLocaltcp(t *testing.T) {
	// Test Localtcp Downstream
	client := NewClient()
	client.upstream = fakeUpstream{} // We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	client.DownstreamManager.StartListeners(&client)

	// Downstream did not yet connect, this should result an error
	packet := makeTestPacket()
	packet.DownstreamId = "net#0"
	packet, err := client.DownstreamManager.Do(packet)
	if err == nil {
		t.Errorf("Could not do packet")
	}

	// Connect downstream
	executor := executor.MakeExecutor()
	go executor.StartClient()
	// Rudimentary way to wait for client to connect
	n := 0
	for len(client.DownstreamManager.GetList()) == 1 {
		if n == 10 {
			t.Error("Waiting 1s for downstream to connect, which didnt happen")
			return
		}
		time.Sleep(100 * time.Millisecond)
		n += 1
	}
	// Check if it works
	packet = makeTestPacket()
	packet.DownstreamId = "net#0"
	packet, err = client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet")
	}
	if !strings.Contains(packet.Response["stdout"], "unreal") {
		t.Errorf("Wrong output")
	}
}

type fakeUpstream struct {
}

func (d fakeUpstream) Start() {
}
func (d fakeUpstream) Channel() chan model.Packet {
	return nil
}
func (d fakeUpstream) SendPacket(packet model.Packet) error {
	return nil
}

func makeTestPacket() model.Packet {
	arguments := make(model.PacketArgument)
	arguments["executable"] = "hostname"
	response := make(model.PacketResponse)
	c := model.NewPacket("exec", "23", "42", arguments, response)
	return c
}
