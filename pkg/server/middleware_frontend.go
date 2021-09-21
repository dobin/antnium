package server

import (
	"github.com/dobin/antnium/pkg/model"
)

func (m *Middleware) FrontendAddNewPacket(packet *model.Packet) error {
	// Add to packet DB and get packetInfo
	packetInfo, err := m.packetDb.addFromFrontend(packet)
	if err != nil {
		return err
	}

	// Notify UI immediately (for initial STATE_RECORDED)
	m.channelFrontendSend <- *packetInfo

	// Send to client, if they are connected via Websocket
	m.channelConnectorSend <- packetInfo.Packet

	return nil
}

func (m *Middleware) FrontendGetAllPacket() []*PacketInfo {
	return m.packetDb.getAll()
}

func (m *Middleware) FrontendGetPacketById(computerId string) []PacketInfo {
	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := m.packetDb.getAll()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ComputerId == computerId {
			filteredPackets = append(filteredPackets, *packetInfo)
		}
	}
	return filteredPackets
}

func (m *Middleware) FrontendGetAllClients() []ClientInfo {
	return m.clientInfoDb.getAsList()
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
