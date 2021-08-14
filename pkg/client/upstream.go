package client

import "github.com/dobin/antnium/pkg/model"

type Upstream interface {
	Connect() error
	Start()
	SendOutofband(packet model.Packet) error
	GetPacket() (model.Packet, error)
	Channel() chan model.Packet
}
