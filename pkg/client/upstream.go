package client

import "github.com/dobin/antnium/pkg/model"

type Upstream interface {
	Start()
	SendPacket(packet model.Packet) error
	Channel() chan model.Packet
}
