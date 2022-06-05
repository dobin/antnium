package server

import (
	"io"
	"testing"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
)

// TestServerPacketIdDuplicate checks if server throws an error when adding two packets with same PacketId
func TestServerPacketIdDuplicate(t *testing.T) {
	port, _ := common.FreePort()
	clientId := "clientid-23"
	packetId := "packetid-43"

	// Server
	s := NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true // Test: WS
	go s.Serve()

	packet := makeSimpleTestPacket(clientId, packetId)

	err := s.Middleware.FrontendAddNewPacket(packet, "")
	if err != nil {
		t.Error("1")
	}
	err = s.Middleware.FrontendAddNewPacket(packet, "")
	if err == nil {
		t.Error("2")
	}

}

func TestServerFileDownloadSecureReference(t *testing.T) {
	port, _ := common.FreePort()

	// Server
	s := NewServer("127.0.0.1:" + port)
	go s.Serve()

	// Client related
	config := client.MakeClientConfig()
	campaign := campaign.MakeCampaign()
	campaign.ServerUrl = "http://127.0.0.1:" + port

	// Create filename
	coder := model.MakeCoder(s.Campaign)
	filenameEncrypted, err := coder.EncryptDataB64([]byte("unittest"))
	if err != nil {
		t.Error("Encrypt")
		return
	}

	// Attempt to download the file
	upstreamRest := client.MakeUpstreamRest(&config, &campaign)
	upstreamRest.Connect()
	url := campaign.ServerUrl + "/secure/" + string(filenameEncrypted)
	resp, err := upstreamRest.HttpGet(url)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	// Decode the file
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}
	decoded, err := coder.DecryptB64Zip(b)
	if err != nil {
		t.Errorf("Decrypt: %s", err.Error())
		return
	}

	// Check if it is the right content
	if string(decoded) != "test" {
		t.Errorf("Content: %s\n", decoded)
		return
	}

}
