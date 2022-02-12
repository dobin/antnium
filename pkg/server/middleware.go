package server

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	channelNewPacket  chan PacketInfo
	channelToFrontend chan PacketInfo
}

func MakeMiddleware(channelNewPacket chan PacketInfo, channelToFrontend chan PacketInfo) Middleware {
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()

	middleware := Middleware{
		packetDb,
		clientInfoDb,
		channelNewPacket,
		channelToFrontend,
	}
	return middleware
}

func (m *Middleware) AddPacketInfo(packetInfo *PacketInfo) {
	m.packetDb.add(packetInfo)
}
