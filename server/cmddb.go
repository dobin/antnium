package server

import (
	"fmt"
	"time"

	"github.com/dobin/antnium/model"
)

type CmdDb struct {
	srvCmd []SrvCmd
}

func MakeCmdDb() CmdDb {
	db := CmdDb{
		make([]SrvCmd, 0),
	}
	return db
}

func (db *CmdDb) add(srvCmd SrvCmd) {
	db.srvCmd = append(db.srvCmd, srvCmd)
}

func (db *CmdDb) getAll() []SrvCmd {
	return db.srvCmd
}

func (db *CmdDb) getCommandFor(computerId string) (model.Command, error) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.State != STATE_RECORDED {
			continue
		}
		srvCmdComputerId := srvCmd.Command.GetComputerId()
		if srvCmdComputerId == "0" || srvCmdComputerId == computerId {
			db.srvCmd[i].State = STATE_SENT // FIXME
			db.srvCmd[i].TimeSent = time.Now()
			return srvCmd.Command, nil
		}
	}

	return nil, fmt.Errorf("Nothing found")
}

func (db *CmdDb) update(command model.Command) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.Command.GetPacketId() == command.GetPacketId() {
			db.srvCmd[i].State = STATE_ANSWERED
			db.srvCmd[i].TimeAnswered = time.Now()
			db.srvCmd[i].Command.SetResponse(command.GetResponse())
			db.srvCmd[i].Command.SetComputerId(command.GetComputerId())
		}
	}
}
