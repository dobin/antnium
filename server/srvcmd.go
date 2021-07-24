package server

import (
	"time"

	"github.com/dobin/antnium/model"
)

type BaseState int

const (
	STATE_RECORDED BaseState = iota
	STATE_SENT
	STATE_ANSWERED
)

type SrvCmd struct {
	Command  model.CommandBase
	State    BaseState
	ClientIp string

	TimeRecorded time.Time
	TimeSent     time.Time
	TimeAnswered time.Time
}

func NewSrvCmd(command model.CommandBase, state BaseState) SrvCmd {
	srvCmd := SrvCmd{
		command,
		state,
		"",
		time.Time{},
		time.Time{},
		time.Time{},
	}
	return srvCmd
}
