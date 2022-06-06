package server

import (
	"testing"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/executor"
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
	campaign := campaign.MakeCampaign()
	campaign.ServerUrl = "http://127.0.0.1:" + port

	executor := executor.MakeExecutor(&campaign)
	fileContent, err := executor.SecureFileDownload("unittest", "", "")
	if err != nil {
		t.Errorf("Download: %s", err.Error())
		return
	}

	// Check if it is the right content
	if string(fileContent) != "test" {
		t.Errorf("Content: %s\n", fileContent)
		return
	}

}
