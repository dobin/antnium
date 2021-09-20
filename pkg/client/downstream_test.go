package client

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/wingman"
)

// TestDownstreamClient tests default Downstream: "Client"
func TestDownstreamClient(t *testing.T) {
	t.Parallel()

	client := NewClient()
	packet := makeExecTestPacket()
	packet, err := client.DownstreamManager.DoIncomingPacket(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Incorrect output")
		return
	}

	// make a second time
	packet = makeExecTestPacket()
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Incorrect output")
		return
	}
}

// TestDownstreamLocaltcp tests if localtcp downstream works
func TestDownstreamLocaltcp(t *testing.T) {
	t.Parallel()

	port, _ := common.GetFreePort()
	portDownstream, _ := common.GetFreePort()
	downstreamTcpAddr := "localhost:" + portDownstream

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.DownstreamManager.downstreamLocaltcp.listenAddr = downstreamTcpAddr

	// We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	fakeUpstream := makeFakeUpstream()
	client.UpstreamManager.UpstreamRest = fakeUpstream
	client.UpstreamManager.UpstreamWs = fakeUpstream

	// Start client, and downstream listeners
	client.Start()
	client.DownstreamManager.StartListeners()

	// Downstream did not yet connect, this should result an error
	packet := makeExecTestPacket()
	packet.DownstreamId = "net#0"
	packet, err := client.DownstreamManager.DoIncomingPacket(packet)
	if err == nil {
		t.Errorf("Could do packet with net#0, even though it should not exist")
		return
	}

	// Connect wingman
	wingman := wingman.MakeWingman()
	go wingman.StartWingman(downstreamTcpAddr)

	// Rudimentary way to wait for client to connect
	n := 0
	for len(client.DownstreamManager.downstreamLocaltcp.DownstreamList()) != 1 {
		if n == 30 {
			t.Error("Waiting 3s for tcp downstream to connect, which didnt happen")
			return
		}
		time.Sleep(100 * time.Millisecond)
		n += 1
	}

	// check if we received oob message
	i := 10
	for i > 0 {
		if fakeUpstream.oobPacket == nil {
			time.Sleep(10)
		}
		i -= 1
	}
	if fakeUpstream.oobPacket == nil {
		t.Errorf("No OOB message")
		return
	}
	if fakeUpstream.oobPacket == nil || fakeUpstream.oobPacket.PacketType != "downstreams" {
		t.Errorf("No OOB notification")
		return
	}

	// Check if it works
	packet = makeExecTestPacket()
	packet.DownstreamId = "net#0"
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}

	// Shutdown client
	wingman.Shutdown()

	// Check if error works, as client is not connected anymore
	packet = makeExecTestPacket()
	packet.DownstreamId = "net#0"
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err == nil {
		t.Errorf("Could do packet")
		return
	}
}

func TestDownstreamDoManager(t *testing.T) {

}

// TestDownstreamLocaltcpRestart tests if the downstream servers survives a restart
func TestDownstreamLocaltcpRestart(t *testing.T) {
	t.Parallel()

	var err error
	port, _ := common.GetFreePort()
	portDownstream, _ := common.GetFreePort()
	downstreamTcpAddr := "localhost:" + portDownstream

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.DownstreamManager.downstreamLocaltcp.listenAddr = downstreamTcpAddr

	// We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	fakeUpstream := makeFakeUpstream()
	client.UpstreamManager.UpstreamRest = fakeUpstream
	client.UpstreamManager.UpstreamWs = fakeUpstream
	client.Start()

	// Test: 1 connected
	if len(client.DownstreamManager.DownstreamServers()) != 1 {
		t.Error("1")
		return
	}

	// Test: Start DownstreamServer
	_, err = client.DownstreamManager.StartListeners()
	if err != nil {
		t.Error(err.Error())
		return
	}

	// Test: 2 connected
	if len(client.DownstreamManager.DownstreamServers()) != 2 {
		t.Error("2")
		return
	}

	// Test: Client Connect ?

	// Test: Shutdown
	client.DownstreamManager.StopListeners()
	// Test: 1 connected
	if len(client.DownstreamManager.DownstreamServers()) != 1 {
		t.Error("3")
		return
	}

	// Test: Client not connect ?

	// Test: Start DownstreamServer
	_, err = client.DownstreamManager.StartListeners()
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Test: 2 connected
	if len(client.DownstreamManager.DownstreamServers()) != 2 {
		t.Error("4")
		return
	}

	// Test: Exec
	// Connect downstream
	wingman0 := wingman.MakeWingman()
	go wingman0.StartWingman(downstreamTcpAddr)
	wingman1 := wingman.MakeWingman()
	go wingman1.StartWingman(downstreamTcpAddr)
	// Rudimentary way to wait for client to connect
	n := 0
	for len(client.DownstreamManager.downstreamLocaltcp.DownstreamList()) != 2 {
		if n == 10 {
			t.Error("Waiting 1s for tcp downstream to connect, which didnt happen")
			return
		}
		time.Sleep(100 * time.Millisecond)
		n += 1
	}

	// check if we received oob message
	i := 10
	for i > 0 {
		if fakeUpstream.oobPacket == nil {
			time.Sleep(10)
		}
		i -= 1
	}
	if fakeUpstream.oobPacket == nil {
		t.Errorf("No OOB message")
		return
	}
	if fakeUpstream.oobPacket == nil || fakeUpstream.oobPacket.PacketType != "downstreams" {
		t.Errorf("No OOB notification")
		return
	}

	// Check if upstream0 it works
	packet := makeExecTestPacket()
	packet.DownstreamId = "net#0"
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}

	// Check if upstream1 it works
	packet = makeExecTestPacket()
	packet.DownstreamId = "net#1"
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}

	// Check if upstream3 does not work
	packet = makeExecTestPacket()
	packet.DownstreamId = "net#3"
	packet, err = client.DownstreamManager.DoIncomingPacket(packet)
	if err == nil {
		t.Errorf("Could do packet: %v", packet)
		return
	}
}

type fakeUpstream struct {
	oobPacket    *model.Packet
	chanIncoming chan model.Packet
	chanOutgoing chan model.Packet
}

func makeFakeUpstream() *fakeUpstream {
	f := fakeUpstream{
		nil,
		make(chan model.Packet),
		make(chan model.Packet),
	}
	return &f
}

func (d *fakeUpstream) Start() {
	/*
		// Collect packets sent to client
		go func() {
			for {
				_ = <-d.chanIncoming
				//p := <-d.chanIncoming
				//d.oobPacket = &p
			}
		}()
	*/

	// Collect packets sent by client
	go func() {
		for {
			p := <-d.chanOutgoing
			d.oobPacket = &p
		}
	}()
}
func (d *fakeUpstream) Connect() error {
	return nil
}
func (d *fakeUpstream) ChanIncoming() chan model.Packet {
	return d.chanIncoming
}
func (d *fakeUpstream) ChanOutgoing() chan model.Packet {
	return d.chanOutgoing
}
func (d *fakeUpstream) Connected() bool {
	return true
}
func (d *fakeUpstream) GetPacket() (model.Packet, error) {
	return model.Packet{}, nil
}

func makeExecTestPacket() model.Packet {
	arguments := make(model.PacketArgument)

	if runtime.GOOS == "windows" {
		arguments["shelltype"] = "cmd"
	} else if runtime.GOOS == "linux" {
		arguments["shelltype"] = "bash"
	} else if runtime.GOOS == "darwin" {
		arguments["shelltype"] = "zsh"
	}

	arguments["commandline"] = "echo test"
	response := make(model.PacketResponse)
	c := model.NewPacket("exec", "23", "42", arguments, response)
	c.DownstreamId = "client"
	return c
}
