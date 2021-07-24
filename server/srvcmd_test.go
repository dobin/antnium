package server

import (
	"encoding/json"
	"testing"

	"github.com/dobin/antnium/model"
)

func TestToJson(t *testing.T) {
	arguments := make(model.CmdArgument)
	arguments["arg0"] = "value0"
	response := make(model.CmdResponse)
	command := model.NewCommand("test", "23", "42", arguments, response)
	srvCmd := NewSrvCmd(command, STATE_RECORDED)

	reference := `{"Command":{"computerid":"23","packetid":"42","command":"test","arguments":{"arg0":"value0"},"response":{}},"State":0,"ClientIp":"","TimeRecorded":"0001-01-01T00:00:00Z","TimeSent":"0001-01-01T00:00:00Z","TimeAnswered":"0001-01-01T00:00:00Z"}`
	u, err := json.Marshal(srvCmd)
	if err != nil {
		panic(err)
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error comparing with reference: " + s)
	}
}

func TestToJsonCommand(t *testing.T) {
	arguments := make(model.CmdArgument)
	arguments["arg0"] = "value0"
	response := make(model.CmdResponse)
	c := model.NewCommand("test", "23", "42", arguments, response)
	reference := `{"computerid":"23","packetid":"42","command":"test","arguments":{"arg0":"value0"},"response":{}}`

	srvCmd := NewSrvCmd(c, STATE_RECORDED)

	u, err := json.Marshal(srvCmd.Command)
	if err != nil {
		panic(err)
	}
	s := string(u)
	if s != reference {
		t.Errorf("Error comparing with reference: " + s)
	}
}
