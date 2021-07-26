package server

import (
	"testing"

	"github.com/dobin/antnium/model"
)

func TestPacketDb(t *testing.T) {
	packetDb := MakePacketDb()

	// Add one Packet to the DB
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	c := model.NewCommand("test", "23", "42", arguments, response)
	packetInfo := NewPacketInfo(c, STATE_RECORDED)
	packetDb.add(packetInfo)

	// Get all packets
	packetInfoAll := packetDb.getAll()
	if len(packetInfoAll) != 1 {
		t.Errorf("Error len packetInfoAll")
	}
	if packetInfoAll[0].State != STATE_RECORDED {
		t.Errorf("Error not right state 1")
	}

	// Client Packet: Should not exist
	_, err := packetDb.getCommandFor("xxx")
	if err == nil {
		t.Errorf("Error packetInfoNotExisting")
	}

	// Client Packet: Should exist
	packetInfoExisting, err := packetDb.getCommandFor("23")
	if err != nil {
		t.Errorf("Error packetInfoExisting 1")
	}
	if packetInfoExisting.Command.ComputerId != "23" {
		t.Errorf("Error packetInfoExisting 2")
	}

	// Client: Again, queue empty
	_, err = packetDb.getCommandFor("23")
	if err == nil {
		t.Errorf("Error packetInfoExisting 11")
	}

	// Backend: Check if exist and right state
	packetInfoAll = packetDb.getAll()
	if packetInfoAll[0].State != STATE_SENT {
		t.Errorf("Error not right state 2")
	}

	// add response from client
	c.Response["ret"] = "oki"
	packetDb.update(c)

	// Server: Should be right state
	packetInfoAll = packetDb.getAll()
	if packetInfoAll[0].State != STATE_ANSWERED {
		t.Errorf("Error not right state 3")
	}
	if packetInfoAll[0].Command.Response["ret"] != "oki" {
		t.Errorf("Error  4")
	}

	// Get the packet for our packet id
	packetInfo, err = packetDb.ByPacketId("42")
	if err != nil {
		t.Errorf("Error  5")
	}
	if packetInfo.Command.ComputerId != "23" {
		t.Errorf("Error  6")
	}
}
