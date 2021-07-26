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

type PacketInfo struct {
	Command  model.Packet
	State    BaseState
	ClientIp string

	TimeRecorded time.Time
	TimeSent     time.Time
	TimeAnswered time.Time
}

func NewPacketInfo(command model.Packet, state BaseState) PacketInfo {
	packetInfo := PacketInfo{
		command,
		state,
		"",
		time.Time{},
		time.Time{},
		time.Time{},
	}
	return packetInfo
}
