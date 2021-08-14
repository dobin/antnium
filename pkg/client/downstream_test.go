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
	port := "50010"
	downstreamTcpAddr := "localhost:50011"

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.DownstreamManager.downstreamLocaltcp.listenAddr = downstreamTcpAddr

	fakeUpstream := fakeUpstream{}
	client.Upstream = &fakeUpstream // We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	client.DownstreamManager.StartListeners(&client)

	// Downstream did not yet connect, this should result an error
	packet := makeTestPacket()
	packet.DownstreamId = "net#0"
	packet, err := client.DownstreamManager.Do(packet)
	if err == nil {
		t.Errorf("Could do packet with net#0, even though it should not exist")
		return
	}

	// Connect downstream
	executor := executor.MakeExecutor()
	go executor.StartClient(downstreamTcpAddr)
	// Rudimentary way to wait for client to connect
	n := 0
	for len(client.DownstreamManager.downstreamLocaltcp.DownstreamList()) != 1 {
		if n == 10 {
			t.Error("Waiting 1s for tcp downstream to connect, which didnt happen")
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
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "unreal") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}
}

type fakeUpstream struct {
}

func (d fakeUpstream) Start() {
}
func (d fakeUpstream) Connect() error {
	return nil
}
func (d fakeUpstream) Channel() chan model.Packet {
	return nil
}
func (d fakeUpstream) SendOutofband(packet model.Packet) error {
	return nil
}
func (d fakeUpstream) GetPacket() (model.Packet, error) {
	return model.Packet{}, nil
}

func makeTestPacket() model.Packet {
	arguments := make(model.PacketArgument)
	arguments["executable"] = "hostname"
	response := make(model.PacketResponse)
	c := model.NewPacket("exec", "23", "42", arguments, response)
	return c
}
