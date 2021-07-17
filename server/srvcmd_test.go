package server

import (
	"encoding/json"
	"testing"

	"github.com/dobin/antnium/model"
)

func TestToJson(t *testing.T) {
	c := model.NewCommandTest("23", "42", []string{"arg0", "arg1"}, "")
	reference := `{"Command":{"computerid":"23","packetid":"42","command":"test","response":"","arguments":["arg0","arg1"]},"State":0,"Source":0}`

	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)

	u, err := json.Marshal(srvCmd)
	if err != nil {
		panic(err)
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error jsonify: " + s)
	}
}

func TestToJsonCommand(t *testing.T) {
	c := model.NewCommandTest("23", "42", []string{"arg0", "arg1"}, "")
	reference := `{"computerid":"23","packetid":"42","command":"test","response":"","arguments":["arg0","arg1"]}`

	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)

	u, err := json.Marshal(srvCmd.Command)
	if err != nil {
		panic(err)
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error jsonify: " + s)
	}
}
