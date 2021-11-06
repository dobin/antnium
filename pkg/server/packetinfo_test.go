package server

import (
	"encoding/json"
	"testing"

	"github.com/dobin/antnium/pkg/model"
)

func TestToJson(t *testing.T) {
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	packet := model.NewPacket("test", "23", "42", arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	packetInfo.User = "user"

	reference := `{"Packet":{"clientid":"23","packetid":"42","packetType":"test","arguments":{"arg0":"value0"},"response":{},"downstreamId":"client"},"State":0,"User":"user","TimeRecorded":"0001-01-01T00:00:00Z","TimeSent":"0001-01-01T00:00:00Z","TimeAnswered":"0001-01-01T00:00:00Z"}`
	u, err := json.Marshal(packetInfo)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error comparing with reference: " + s)
		return
	}
}

func TestToJsonPacket(t *testing.T) {
	arguments := make(model.PacketArgument)
	arguments["arg0"] = "value0"
	response := make(model.PacketResponse)
	c := model.NewPacket("test", "23", "42", arguments, response)
	reference := `{"clientid":"23","packetid":"42","packetType":"test","arguments":{"arg0":"value0"},"response":{},"downstreamId":"client"}`

	packetInfo := NewPacketInfo(c, STATE_RECORDED)

	u, err := json.Marshal(packetInfo.Packet)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error comparing with reference: " + s)
	}
}
