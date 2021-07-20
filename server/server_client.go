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
	s.hostDb.updateFor(computerId)

	srvCmd, err := s.cmdDb.getCommandFor(computerId)
	if err != nil {
		return
	}

	// Set source IP for this command
	srvCmd.ClientIp = r.RemoteAddr

	// Encode the command and send it
	jsonData, err := s.coder.EncodeData(srvCmd.Command)
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"command": srvCmd.Command,
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
		return
	}
	log.WithFields(log.Fields{
		"command": command,
	}).Info("Send command")
	s.cmdDb.update(command)
	s.hostDb.updateFor(command.ComputerId)
	fmt.Fprint(rw, "asdf")
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

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
