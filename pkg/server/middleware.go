package server

type Middleware struct {
	packetDb     PacketDb
	clientInfoDb ClientInfoDb

	connectorManager *ConnectorManager
	frontendManager  *FrontendManager
}

func MakeMiddleware() Middleware {
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()

	middleware := Middleware{
		packetDb,
		clientInfoDb,
		nil,
		nil,
	}
	return middleware
}

func (m *Middleware) SetTODO(connectorManager *ConnectorManager, frontendManager *FrontendManager) {
	m.connectorManager = connectorManager
	m.frontendManager = frontendManager
}
