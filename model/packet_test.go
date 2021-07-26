package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFromJson(t *testing.T) {
	a := `{"computerid":"23","packetid":"42","command":"test","arguments":{"arg0":"value0"},"response":{"foo":"bar"}}`
	var packet Packet
	err := json.Unmarshal([]byte(a), &packet)
	if err != nil {
		t.Errorf("Could not parse packet test 1: %s", err)
	}
	if packet.Response["foo"] != "bar" {
		t.Errorf("Could not parse packet test 2: %s", err)
	}
}

func TestToJson(t *testing.T) {
	arguments := make(PacketArgument)
	arguments["arg0"] = "value0"
	response := make(PacketResponse)
	c := NewPacket("test", "23", "42", arguments, response)

	reference := `{"computerid":"23","packetid":"42","command":"test","arguments":{"arg0":"value0"},"response":{}}`
	u, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	s := string(u)
	fmt.Println(s)
	if s != reference {
		t.Errorf("Error jsonify: " + s)
	}
}
