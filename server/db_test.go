package server

import (
	"testing"

	"github.com/dobin/antnium/model"
)

func TestDb(t *testing.T) {
	db := MakeDb()

	c := model.NewCommandTest("23", "42", []string{"arg0", "arg1"}, "")
	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)

	db.add(srvCmd)

	srvCmdAll := db.getAll()
	if len(srvCmdAll) != 1 {
		t.Errorf("Error len srvCmdAll")
	}
	if srvCmdAll[0].State != STATE_RECORDED {
		t.Errorf("Error not right state 1")
	}

	// Client: Should not exist
	_, err := db.getCommandFor("xxx")
	if err == nil {
		t.Errorf("Error srvCmdNotExisting")
	}

	// Client: Should exist
	srvCmdExisting, err := db.getCommandFor("23")
	if err != nil {
		t.Errorf("Error srvCmdExisting 1")
	}
	if srvCmdExisting.GetComputerId() != "23" {
		t.Errorf("Error srvCmdExisting 2")
	}

	// Client: Again, queue empty
	_, err = db.getCommandFor("23")
	if err == nil {
		t.Errorf("Error srvCmdExisting 11")
	}

	// Backend: Check if exist and right state
	srvCmdAll = db.getAll()
	if srvCmdAll[0].State != STATE_SENT {
		t.Errorf("Error not right state 2")
	}

	// add response from client
	c.Response = "oki"
	db.update(c)

	// Server: Should be right state
	srvCmdAll = db.getAll()
	if srvCmdAll[0].State != STATE_ANSWERED {
		t.Errorf("Error not right state 3")
	}
	if srvCmdAll[0].Command.GetResponse() != "oki" {
		t.Errorf("Error  4")
	}

}
