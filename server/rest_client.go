package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	// Update last seen for this host
	s.clientInfoDb.updateFor(computerId, r.RemoteAddr)

	packetInfo, err := s.packetDb.getCommandFor(computerId)
	if err != nil {
		return
	}

	// only notify if we had a command for the client
	s.adminWebSocket.broadcastPacket(*packetInfo)

	// Set source IP for this command
	packetInfo.ClientIp = r.RemoteAddr

	// Encode the command and send it
	jsonData, err := s.coder.EncodeData(packetInfo.Command)
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"command": packetInfo.Command,
	}).Info("Get command")
	fmt.Fprint(rw, string(jsonData))
}

func (s *Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}
	command, err := s.coder.DecodeData(reqBody)
	if err != nil {
		log.Error("Could not decode")
		return
	}
	log.WithFields(log.Fields{
		"command": command,
	}).Info("Send command")

	s.clientInfoDb.updateFor(command.ComputerId, r.RemoteAddr)

	packetInfo, err := s.packetDb.update(command)
	if err == nil {
		// only broadcast if element has been found (against ping-packet spam)
		s.adminWebSocket.broadcastPacket(packetInfo)
	}

	fmt.Fprint(rw, "asdf")
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

	// Check if request for this file really exists
	packetInfo, err := s.packetDb.ByPacketId(packetId)
	if err != nil {
		log.Warnf("Client attempted to upload a file with an expired command with packetid: %s: %s",
			packetId, err.Error())
		return
	}
	if packetInfo.State != STATE_SENT {
		log.Warnf("Client attempted to upload a file with an weird command state %d",
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
