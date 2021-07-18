package server

import (
	"fmt"

	"github.com/dobin/antnium/model"
)

type Db struct {
	srvCmd []SrvCmd
}

func MakeDb() Db {
	db := Db{
		make([]SrvCmd, 0),
	}
	return db
}

func (db *Db) add(srvCmd SrvCmd) {
	db.srvCmd = append(db.srvCmd, srvCmd)
}

func (db *Db) getAll() []SrvCmd {
	return db.srvCmd
}

func (db *Db) getCommandFor(computerId string) (model.Command, error) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.State != STATE_RECORDED {
			continue
		}
		srvCmdComputerId := srvCmd.Command.GetComputerId()
		if srvCmdComputerId == "0" || srvCmdComputerId == computerId {
			db.srvCmd[i].State = STATE_SENT // FIXME
			return srvCmd.Command, nil
		}
	}

	return nil, fmt.Errorf("Nothing found")
}

func (db *Db) update(command model.Command) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.Command.GetPacketId() == command.GetPacketId() {
			db.srvCmd[i].Command.SetResponse(command.GetResponse())
			db.srvCmd[i].State = STATE_ANSWERED
			db.srvCmd[i].Command.SetComputerId(command.GetComputerId())
		}
	}
}
