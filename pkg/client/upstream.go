package client

import (
	"github.com/dobin/antnium/pkg/model"
)

type Upstream interface {
	Connect() error
	Connected() bool
	Start()
	ChanIncoming() chan model.Packet
	SendPacket(packet model.Packet) error
}
