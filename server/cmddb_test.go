package server

import (
	"testing"

	"github.com/dobin/antnium/model"
)

func TestDb(t *testing.T) {
	cmdDb := MakeCmdDb()

	arguments := make(model.CmdArgument)
	arguments["arg0"] = "value0"
	response := make(model.CmdResponse)
	c := model.NewCommand("test", "23", "42", arguments, response)
	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)

	cmdDb.add(srvCmd)

	srvCmdAll := cmdDb.getAll()
	if len(srvCmdAll) != 1 {
		t.Errorf("Error len srvCmdAll")
	}
	if srvCmdAll[0].State != STATE_RECORDED {
		t.Errorf("Error not right state 1")
	}

	// Client: Should not exist
	_, err := cmdDb.getCommandFor("xxx")
	if err == nil {
		t.Errorf("Error srvCmdNotExisting")
	}

	// Client: Should exist
	srvCmdExisting, err := cmdDb.getCommandFor("23")
	if err != nil {
		t.Errorf("Error srvCmdExisting 1")
	}
	if srvCmdExisting.ComputerId != "23" {
		t.Errorf("Error srvCmdExisting 2")
	}

	// Client: Again, queue empty
	_, err = cmdDb.getCommandFor("23")
	if err == nil {
		t.Errorf("Error srvCmdExisting 11")
	}

	// Backend: Check if exist and right state
	srvCmdAll = cmdDb.getAll()
	if srvCmdAll[0].State != STATE_SENT {
		t.Errorf("Error not right state 2")
	}

	// add response from client
	c.Response["ret"] = "oki"
	cmdDb.update(c)

	// Server: Should be right state
	srvCmdAll = cmdDb.getAll()
	if srvCmdAll[0].State != STATE_ANSWERED {
		t.Errorf("Error not right state 3")
	}
	if srvCmdAll[0].Command.Response["ret"] != "oki" {
		t.Errorf("Error  4")
	}

}
