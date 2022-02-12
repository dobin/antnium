package server

import (
	"fmt"
	"io"
	"os"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (m *Middleware) FrontendAddNewPacket(packet *model.Packet, user string) error {
	// Add to packet DB and get packetInfo
	packetInfo, err := m.packetDb.addFromFrontend(packet, user)
	if err != nil {
		return err
	}

	// Notify UI immediately (for initial STATE_RECORDED)
	m.channelToFrontend <- *packetInfo

	// Send to client, if they are connected via Websocket
	m.channelToClients <- *packetInfo

	return nil
}

func (m *Middleware) FrontendAllPacket() []*PacketInfo {
	return m.packetDb.All()
}

func (m *Middleware) FrontendGetPacketById(clientId string) []PacketInfo {
	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := m.packetDb.All()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ClientId == clientId {
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

func (m *Middleware) AdminUploadFile(basename string, httpFile io.ReadCloser) error {
	filename := fmt.Sprintf("static/%s", basename)

	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("destination file %s already exists", filename)
	}

	out, err := os.Create(filename)
	if err != nil {
		log.Error("Middleware: AdminUploadFile: Could not open file: " + filename)
		return err
	}
	defer out.Close()

	written, err := io.Copy(out, httpFile)
	if err != nil {
		log.Error("Middleware: AdminUploadFile: Error copying: " + err.Error())
		return err
	}

	log.Infof("Middleware: AdminUploadFile: Written %d bytes to file %s", written, filename)
	return nil
}
