package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dobin/antnium/model"
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

	// only notify if we had a packet for the client
	s.adminWebSocket.broadcastPacket(*packetInfo)

	// Set source IP for this packet
	packetInfo.ClientIp = r.RemoteAddr

	// Encode the packet and send it
	jsonData, err := s.coder.EncodeData(packetInfo.Packet)
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"packet": packetInfo.Packet,
	}).Info("Get packet")
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
	log.WithFields(log.Fields{
		"packet": packet,
	}).Info("Send packet")

	s.clientInfoDb.updateFor(packet.ComputerId, r.RemoteAddr)

	if packet.PacketType == "ping" {
		s.handlePingPacket(packet)
	}

	packetInfo, err := s.packetDb.update(packet)
	if err == nil {
		// only broadcast if element has been found (against ping-packet spam)
		s.adminWebSocket.broadcastPacket(packetInfo)
	}

	fmt.Fprint(rw, "asdf")
}

func (s *Server) handlePingPacket(packet model.Packet) {

	hostname, _ := packet.Response["hostname"]
	localIps := model.ResponseToArray("localIp", packet.Response)

	s.clientInfoDb.updateMore(packet.ComputerId, hostname, localIps)
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

	filename := "upload/" + packetId

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
