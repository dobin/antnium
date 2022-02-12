package server

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	channelToClients  chan PacketInfo
	channelToFrontend chan PacketInfo
}

func MakeMiddleware(channelToClients chan PacketInfo, channelToFrontend chan PacketInfo) Middleware {
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()

	middleware := Middleware{
		packetDb,
		clientInfoDb,
		channelToClients,
		channelToFrontend,
	}
	return middleware
}

func (m *Middleware) AddPacketInfo(packetInfo *PacketInfo) {
	m.packetDb.add(packetInfo)
}
