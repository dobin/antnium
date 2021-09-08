package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (s *Middleware) ClientSendPacket(packet model.Packet, remoteAddr string, connectorType string) {
	if packet.PacketType == "ping" {
		s.clientInfoDb.updateFromPing(packet.ComputerId, remoteAddr, connectorType, packet.Response)
		return
	}

	s.addNewClientPacket(packet, remoteAddr, connectorType)
}

func (s *Middleware) addNewClientPacket(packet model.Packet, remoteAddr string, connectorType string) {
	// Update Client DB
	s.clientInfoDb.updateFor(packet.ComputerId, remoteAddr, connectorType)

	// Update Package DB
	packetInfo := s.packetDb.update(packet)

	// Notify UI
	s.frontendManager.FrontendWs.broadcastPacket(packetInfo)
}

func (s *Middleware) ClientGetPacket(computerId string, remoteAddr string, connectorType string) (model.Packet, bool) {
	// Update last seen for this host
	s.clientInfoDb.updateFor(computerId, remoteAddr, connectorType)

	packetInfo, err := s.packetDb.getPacketFor(computerId)
	if err != nil {
		return model.Packet{}, false
	}

	// only notify UI if we really sent a packet
	s.frontendManager.FrontendWs.broadcastPacket(*packetInfo)

	// Set source IP for this packet
	packetInfo.ClientIp = remoteAddr

	return packetInfo.Packet, true
}

func (s *Middleware) ClientUploadFile(packetId string, httpFile io.ReadCloser) {
	// Check if request for this file really exists
	packetInfo, err := s.packetDb.ByPacketId(packetId)
	if err != nil {
		log.Warnf("Client attempted to upload a file with an expired packet with packetid: %s: %s",
			packetId, err.Error())
		return
	}
	if packetInfo.State != STATE_SENT {
		log.Warnf("Client attempted to upload a file with an weird packet state %d",
			packetInfo.State)
		return
	}

	basename := filepath.Base(packetInfo.Packet.Arguments["source"])
	filename := fmt.Sprintf("upload/%s.%s.%s",
		packetInfo.Packet.ComputerId,
		packetInfo.Packet.PacketId,
		basename,
	)

	out, err := os.Create(filename)
	if err != nil {
		log.Error("Could not open file: " + filename)
		return
	}
	defer out.Close()

	written, err := io.Copy(out, httpFile)
	if err != nil {
		log.Error("Error copying: " + err.Error())
		return
	}

	log.Infof("Written %d bytes to file %s", written, packetId)
}
