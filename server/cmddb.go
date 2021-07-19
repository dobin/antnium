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
	srvCmd.TimeRecorded = time.Now()
	db.srvCmd = append(db.srvCmd, srvCmd)
}

func (db *CmdDb) getAll() []SrvCmd {
	return db.srvCmd
}

func (db *CmdDb) getCommandFor(computerId string) (*SrvCmd, error) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.State != STATE_RECORDED {
			continue
		}
		srvCmdComputerId := srvCmd.Command.ComputerId
		if srvCmdComputerId == "0" || srvCmdComputerId == computerId {
			db.srvCmd[i].State = STATE_SENT // FIXME
			db.srvCmd[i].TimeSent = time.Now()
			return &db.srvCmd[i], nil
		}
	}

	return &SrvCmd{}, fmt.Errorf("Nothing found")
}

func (db *CmdDb) update(command model.CommandBase) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.Command.PacketId == command.PacketId {
			db.srvCmd[i].State = STATE_ANSWERED
			db.srvCmd[i].TimeAnswered = time.Now()
			db.srvCmd[i].Command.Response = command.Response
			db.srvCmd[i].Command.ComputerId = command.ComputerId
		}
	}
}
