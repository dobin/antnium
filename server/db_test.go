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

	srvCmdNotExisting := db.getCommandsFor("xxx")
	if len(srvCmdNotExisting) != 0 {
		t.Errorf("Error len srvCmdNotExisting: %d", len(srvCmdNotExisting))
	}

	srvCmdExisting := db.getCommandsFor("23")
	if len(srvCmdExisting) != 1 {
		t.Errorf("Error len srvCmdExisting: %d", len(srvCmdExisting))
	}

	srvCmdAll = db.getAll()
	if srvCmdAll[0].State != STATE_SENT {
		t.Errorf("Error not right state 2")
	}

	c.Response = "oki"
	db.update(c)

	srvCmdAll = db.getAll()
	if srvCmdAll[0].State != STATE_ANSWERED {
		t.Errorf("Error not right state 3")
	}
	if srvCmdAll[0].Command.GetResponse() != "oki" {
		t.Errorf("Error  4")
	}

}
