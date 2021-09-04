package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// getPacket provides a client with new packets, if any
func (s *HttpServer) getPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	if computerId == "" {
		return
	}

	packet, ok := s.serverManager.ClientGetPacket(computerId, r.RemoteAddr)
	if !ok {
		// No packet, just return
		return
	}

	// Encode the packet and send it
	jsonData, err := s.coder.EncodeData(packet)
	if err != nil {
		return
	}

	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
	}).Info("ToClient   ")

	fmt.Fprint(rw, string(jsonData))
}

// sendPacket receives packet answers from client
func (s *HttpServer) sendPacket(rw http.ResponseWriter, r *http.Request) {
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

	s.serverManager.ClientSendPacket(packet, r.RemoteAddr)

	fmt.Fprint(rw, "asdf")
}

func (s *HttpServer) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

	s.serverManager.ClientUploadFile(packetId, r.Body)

	fmt.Fprintf(w, "ok\n")
}
