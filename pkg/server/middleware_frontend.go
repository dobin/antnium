package server

import (
	"github.com/dobin/antnium/pkg/model"
)

func (m *Middleware) FrontendAddNewPacket(packet *model.Packet, user string) error {
	// Add to packet DB and get packetInfo
	packetInfo, err := m.packetDb.addFromFrontend(packet, user)
	if err != nil {
		return err
	}

	// Notify UI immediately (for initial STATE_RECORDED)
	m.frontendSend <- *packetInfo

	// Send to client, if they are connected via Websocket
	m.connectorSend <- packetInfo.Packet

	return nil
}

func (m *Middleware) FrontendAllPacket() []*PacketInfo {
	return m.packetDb.All()
}

func (m *Middleware) FrontendGetPacketById(computerId string) []PacketInfo {
	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := m.packetDb.All()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ComputerId == computerId {
			filteredPackets = append(filteredPackets, *packetInfo)
		}
	}
	return filteredPackets
}

func (m *Middleware) FrontendAllClients() []ClientInfo {
	return m.clientInfoDb.AllAsList()
}

/*
func (m *Middleware) FrontendGetCampaign() campaign.Campaign {
	return *m.Campaign
}

func (m *Middleware) FrontendDirUpload() []model.DirEntry {
}

func (m *Middleware) FrontendDirStatic() {
}
*/
