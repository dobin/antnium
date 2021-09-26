package server

import (
	"time"

	"github.com/dobin/antnium/pkg/model"
)

type BaseState int

const (
	STATE_RECORDED BaseState = iota
	STATE_SENT
	STATE_ANSWERED
	STATE_CLIENT
)

type PacketInfo struct {
	Packet model.Packet
	State  BaseState
	User   string

	TimeRecorded time.Time
	TimeSent     time.Time
	TimeAnswered time.Time
}

func NewPacketInfo(packet model.Packet, state BaseState) PacketInfo {
	packetInfo := PacketInfo{
		packet,
		state,
		"",
		time.Time{},
		time.Time{},
		time.Time{},
	}
	return packetInfo
}
