package server

import (
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/websocket"
)

func (s *ServerManager) AdminGetAllPacket() []PacketInfo {
	return s.packetDb.getAll()
}

func (s *ServerManager) AdminGetPacketById(computerId string) []PacketInfo {
	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := s.packetDb.getAll()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ComputerId == computerId {
			filteredPackets = append(filteredPackets, packetInfo)
		}
	}
	return filteredPackets
}

func (s *ServerManager) AdminGetAllClients() []ClientInfo {
	return s.clientInfoDb.getAsList()
}

func (s *ServerManager) AdminAddNewPacket(packet model.Packet) {
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)

	// Add to DB and get updated one
	packetInfo = s.packetDb.add(packetInfo)

	// Notify UI immediately (for initial STATE_RECORDED)
	s.adminWebSocket.broadcastPacket(packetInfo)

	// Send to client, if they are connected via Websocket
	ok := s.clientWebSocket.TryNotify(&packetInfo.Packet)
	if ok {
		// only notify UI if we really sent a packet
		s.adminWebSocket.broadcastPacket(packetInfo)
	}
}

/*
func (s *ServerManager) AdminGetCampaign() campaign.Campaign {
	return *s.Campaign
}

func (s *ServerManager) AdminDirUpload() []model.DirEntry {
}

func (s *ServerManager) AdminDirStatic() {
}
*/

func (s *ServerManager) AdminRegisterWs(wsConn *websocket.Conn) {
	s.adminWebSocket.registerWs(wsConn)
}

func (s *ServerManager) ClientRegisterWs(computerId string, wsConn *websocket.Conn) {
	s.clientWebSocket.registerWs(computerId, wsConn)
}
