package server

import (
	"testing"

	"github.com/dobin/antnium/pkg/common"
)

// TestServerPacketIdDuplicate checks if server throws an error when adding two packets with same PacketId
func TestServerPacketIdDuplicate(t *testing.T) {
	port, _ := common.GetFreePort()
	computerId := "computerid-23"
	packetId := "packetid-43"

	// Server
	s := NewServer("127.0.0.1:" + port)
	s.Campaign.ClientUseWebsocket = true // Test: WS
	go s.Serve()

	packet := makeSimpleTestPacket(computerId, packetId)

	err := s.Middleware.FrontendAddNewPacket(packet)
	if err != nil {
		t.Error("1")
	}
	err = s.Middleware.FrontendAddNewPacket(packet)
	if err == nil {
		t.Error("2")
	}

}
