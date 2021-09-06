package server

import (
	"github.com/dobin/antnium/pkg/model"
)

func (s *Middleware) AdminAddNewPacket(packet model.Packet) {
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)

	// Add to DB and get updated one
	packetInfo = s.packetDb.add(packetInfo)

	// Notify UI immediately (for initial STATE_RECORDED)
	s.frontendManager.FrontendWs.broadcastPacket(packetInfo)

	// Send to client, if they are connected via Websocket
	ok := s.connectorManager.ConnectorWs.TryViaWebsocket(&packetInfo.Packet)
	if ok {
		// only notify UI if we really sent a packet
		s.frontendManager.FrontendWs.broadcastPacket(packetInfo)
	}
}

func (s *Middleware) AdminGetAllPacket() []PacketInfo {
	return s.packetDb.getAll()
}

func (s *Middleware) AdminGetPacketById(computerId string) []PacketInfo {
	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := s.packetDb.getAll()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ComputerId == computerId {
			filteredPackets = append(filteredPackets, packetInfo)
		}
	}
	return filteredPackets
}

func (s *Middleware) AdminGetAllClients() []ClientInfo {
	return s.clientInfoDb.getAsList()
}

/*
func (s *Middleware) AdminGetCampaign() campaign.Campaign {
	return *s.Campaign
}

func (s *Middleware) AdminDirUpload() []model.DirEntry {
}

func (s *Middleware) AdminDirStatic() {
}
*/
