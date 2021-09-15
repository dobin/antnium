package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

// ClientSendPacket handles packets sent by client (either answers, or client-initiated)
func (s *Middleware) ClientSendPacket(packet model.Packet, remoteAddr string, connectorType string) error {
	common.LogPacketDebug("Server:ClientSendPacket()", packet)

	// Update Client DB
	s.clientInfoDb.updateFor(packet.ComputerId, remoteAddr, connectorType)

	// Special care packets
	if packet.PacketType == "ping" {
		// Dont store it, clientInfoDb update was enough
		return nil
	} else if packet.PacketType == "clientinfo" {
		// Update our DB with client info data.
		// But still store it, and broadcast it (used in the frontend to detect newly connected clients)
		s.clientInfoDb.updateFromClientinfo(packet.ComputerId, remoteAddr, connectorType, packet.Response)
	}

	// Update Package DB
	packetInfo := s.packetDb.updateFromClient(packet)

	// Notify UI
	s.channelFrontendSend <- *packetInfo

	return nil
}

func (s *Middleware) ClientGetPacket(computerId string, remoteAddr string, connectorType string) (model.Packet, bool) {
	log.Debugf("Server:ClientGetPacket(): %s", computerId)

	// Update last seen for this host
	s.clientInfoDb.updateFor(computerId, remoteAddr, connectorType)

	// Check if we have any packets available
	packetInfo, err := s.packetDb.getPacketForClient(computerId)
	if err != nil {
		return model.Packet{}, false
	}

	// Update packet infos
	s.packetDb.sentToClient(packetInfo.Packet.PacketId, remoteAddr)

	// notify UI about it
	s.channelFrontendSend <- *packetInfo

	return packetInfo.Packet, true
}

func (s *Middleware) ClientUploadFile(packetId string, httpFile io.ReadCloser) {
	// Check if request for this file really exists
	packetInfo, ok := s.packetDb.ByPacketId(packetId)
	if !ok {
		log.Warnf("Client attempted to upload a file with an expired packet with packetid: %s",
			packetId)
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
