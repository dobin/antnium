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
	fmt.Printf("Add SrvCmd: %v\n", srvCmd)
	db.srvCmd = append(db.srvCmd, srvCmd)
}

func (db *Db) getAll() []SrvCmd {
	fmt.Printf("GetAll SrvCmd\n")
	return db.srvCmd
}

func (db *Db) getCommandsFor(computerId string) []model.Command {
	fmt.Printf("GetCommandsFor\n")
	commands := make([]model.Command, 0)

	for i, srvCmd := range db.srvCmd {
		srvCmdComputerId := srvCmd.Command.GetComputerId()
		if srvCmdComputerId == "0" || srvCmdComputerId == computerId {
			db.srvCmd[i].State = STATE_SENT // FIXME
			commands = append(commands, srvCmd.Command)
		}
	}

	return commands
}

func (db *Db) update(command model.Command) {
	for i, srvCmd := range db.srvCmd {
		if srvCmd.Command.GetPacketId() == command.GetPacketId() {
			db.srvCmd[i].Command.SetResponse(command.GetResponse())
			db.srvCmd[i].State = STATE_ANSWERED
		}
	}
}
