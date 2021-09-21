package server

import "github.com/dobin/antnium/pkg/model"

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	connectorSend chan model.Packet
	frontendSend  chan PacketInfo
}

func MakeMiddleware(channelConnectorSend chan model.Packet, channelFrontendSend chan PacketInfo) Middleware {
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()

	middleware := Middleware{
		packetDb,
		clientInfoDb,
		channelConnectorSend,
		channelFrontendSend,
	}
	return middleware
}

func (m *Middleware) AddPacketInfo(packetInfo *PacketInfo) {
	m.packetDb.add(packetInfo)
}
