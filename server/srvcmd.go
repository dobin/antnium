package server

import (
	"github.com/dobin/antnium/model"
)

type BaseState int

const (
	STATE_RECORDED BaseState = iota
	STATE_SENT
	STATE_ANSWERED
)

type BaseSource int

const (
	SOURCE_SRV BaseSource = iota
	SOURCE_CLI
)

type SrvCmd struct {
	Command model.Command
	State   BaseState
	Source  BaseSource
}

func NewSrvCmd(command model.Command, state BaseState, source BaseSource) SrvCmd {
	srvCmd := SrvCmd{
		command,
		state,
		source,
	}
	return srvCmd
}
