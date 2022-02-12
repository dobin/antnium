package server

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	connectorSend chan PacketInfo
	frontendSend  chan PacketInfo
}

func MakeMiddleware(channelConnectorSend chan PacketInfo, channelFrontendSend chan PacketInfo) Middleware {
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
