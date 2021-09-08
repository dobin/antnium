package client

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dobin/antnium/pkg/downstreamclient"
	"github.com/dobin/antnium/pkg/model"
)

func TestDownstreamClient(t *testing.T) {
	// Test default Downstream: "Client"
	client := NewClient()
	client.Campaign.ClientUseWebsocket = false
	packet := makeTestPacket()
	packet, err := client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Incorrect output")
		return
	}

	// make a second time
	packet = makeTestPacket()
	packet, err = client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Incorrect output")
		return
	}
}

func TestDownstreamLocaltcp(t *testing.T) {
	port := "50013"
	downstreamTcpAddr := "localhost:50011"

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false
	client.DownstreamManager.downstreamLocaltcp.listenAddr = downstreamTcpAddr

	fakeUpstream := makeFakeUpstream()
	client.UpstreamManager.UpstreamHttp = fakeUpstream // We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	client.UpstreamManager.UpstreamWs = fakeUpstream

	client.Start()
	client.DownstreamManager.StartListeners()

	// Downstream did not yet connect, this should result an error
	packet := makeTestPacket()
	packet.DownstreamId = "net#0"
	packet, err := client.DownstreamManager.Do(packet)
	if err == nil {
		t.Errorf("Could do packet with net#0, even though it should not exist")
		return
	}

	// Connect downstreamClient
	downstreamClient := downstreamclient.MakeDownstreamClient()
	go downstreamClient.StartClient(downstreamTcpAddr)
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
	packet = makeTestPacket()
	packet.DownstreamId = "net#0"
	packet, err = client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}
}

func TestDownstreamDoManager(t *testing.T) {

}

func TestDownstreamLocaltcpRestart(t *testing.T) {
	port := "50013"
	downstreamTcpAddr := "localhost:60000"

	// Test Localtcp Downstream
	client := NewClient()
	client.Campaign.ServerUrl = "http://127.0.0.1:" + port
	client.Campaign.ClientUseWebsocket = false
	client.DownstreamManager.downstreamLocaltcp.listenAddr = downstreamTcpAddr

	fakeUpstream := makeFakeUpstream()
	client.UpstreamManager.UpstreamHttp = fakeUpstream // We dont have an upstream, so fake one so we dont do HTTP requests to nowhere
	client.UpstreamManager.UpstreamWs = fakeUpstream
	client.Start()

	var err error

	// Test: Server list 1
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
	// Test: Server list 2
	if len(client.DownstreamManager.DownstreamServers()) != 2 {
		t.Error("2")
		return
	}
	// Test: Client Connect ?

	// Test: Shutdown
	client.DownstreamManager.StopListeners()
	// Test: Server list 1
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
	// Test: Server list 2
	if len(client.DownstreamManager.DownstreamServers()) != 2 {
		t.Error("4")
		return
	}

	// Test: Exec
	// Connect downstream
	downstreamClient := downstreamclient.MakeDownstreamClient()
	go downstreamClient.StartClient(downstreamTcpAddr)
	downstreamClient2 := downstreamclient.MakeDownstreamClient()
	go downstreamClient2.StartClient(downstreamTcpAddr)
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

	// Check if it works
	packet := makeTestPacket()
	packet.DownstreamId = "net#1"
	packet, err = client.DownstreamManager.Do(packet)
	if err != nil {
		t.Errorf("Could not do packet: %s", err.Error())
		return
	}
	if !strings.Contains(packet.Response["stdout"], "test") {
		t.Errorf("Wrong output, got: %v", packet.Response)
		return
	}

	//t.Error("asdf")
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
	// Collect packets of the upstream
	go func() {
		for {
			p := <-d.chanIncoming
			d.oobPacket = &p
		}
	}()
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

func makeTestPacket() model.Packet {
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
