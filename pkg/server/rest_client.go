package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// getPacket provides a client with new packets, if any
func (s *Server) getPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	// Update last seen for this host
	s.clientInfoDb.updateFor(computerId, r.RemoteAddr)

	packetInfo, err := s.packetDb.getPacketFor(computerId)
	if err != nil {
		return
	}

	// only notify UI if we really sent a packet
	s.adminWebSocket.broadcastPacket(*packetInfo)

	// Set source IP for this packet
	packetInfo.ClientIp = r.RemoteAddr

	// Encode the packet and send it
	jsonData, err := s.coder.EncodeData(packetInfo.Packet)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"1_computerId":   packetInfo.Packet.ComputerId,
		"2_packetId":     packetInfo.Packet.PacketId,
		"3_downstreamId": packetInfo.Packet.DownstreamId,
		"4_packetType":   packetInfo.Packet.PacketType,
		"5_arguments":    packetInfo.Packet.Arguments,
	}).Info("ToClient   ")

	fmt.Fprint(rw, string(jsonData))
}

// sendPacket receives packet answers from client
func (s *Server) sendPacket(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}
	packet, err := s.coder.DecodeData(reqBody)
	if err != nil {
		log.Error("Could not decode")
		return
	}
	/*
		log.WithFields(log.Fields{
			"computerId":   packet.ComputerId,
			"packetId":     packet.PacketId,
			"downstreamId": packet.DownstreamId,
			"packetType":   packet.PacketType,
			"argumetns":    packet.Arguments,
		}).Info("Recv from Client")
	*/
	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Info("FromClient ")

	if packet.PacketType == "ping" {
		s.clientInfoDb.updateFromPing(packet.ComputerId, r.RemoteAddr, packet.Response)
		fmt.Fprint(rw, "asdf")
		return
	}

	s.clientInfoDb.updateFor(packet.ComputerId, r.RemoteAddr)

	packetInfo := s.packetDb.update(packet)
	s.adminWebSocket.broadcastPacket(packetInfo)

	fmt.Fprint(rw, "asdf")
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

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

	written, err := io.Copy(out, r.Body)
	if err != nil {
		log.Error("Error copying: " + err.Error())
		return
	}

	log.Infof("Written %d bytes to file %s", written, packetId)

	fmt.Fprintf(w, "ok\n")
}
