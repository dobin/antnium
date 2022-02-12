package server

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	channelSend chan PacketInfo
}

func MakeMiddleware(channelSend chan PacketInfo) Middleware {
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()

	middleware := Middleware{
		packetDb,
		clientInfoDb,
		channelSend,
	}
	return middleware
}

func (m *Middleware) AddPacketInfo(packetInfo *PacketInfo) {
	m.packetDb.add(packetInfo)
}
