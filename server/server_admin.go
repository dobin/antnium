package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dobin/antnium/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *Server) adminListCommands(rw http.ResponseWriter, r *http.Request) {
	srvCmds := s.cmdDb.getAll()
	json, err := json.Marshal(srvCmds)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminListCommandsComputerId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	var filteredCmds []SrvCmd = make([]SrvCmd, 0)
	srvCmds := s.cmdDb.getAll()
	for _, srvCmd := range srvCmds {
		if srvCmd.Command.ComputerId == computerId {
			filteredCmds = append(filteredCmds, srvCmd)
		}
		if len(srvCmds) > 5 {
			break
		}
	}

	json, err := json.Marshal(filteredCmds)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := s.hostDb.getAsList()
	json, err := json.Marshal(hostList)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminAddTestCommand(rw http.ResponseWriter, r *http.Request) {
	arguments := make(model.CmdArgument)
	//arguments["executable"] = "cmd"
	//arguments["arg1"] = "/C"
	//arguments["arg2"] = "whoami"

	//arguments["remoteurl"] = "http://127.0.0.1:4444/psexec.txt"
	//arguments["destination"] = "psexec.txt"

	packetId := s.getRandomPacketId()

	arguments["remoteurl"] = "http://127.0.0.1:4444/upload/" + packetId
	arguments["source"] = "README.md"

	response := make(model.CmdResponse)
	command := model.NewCommand("fileupload", "0", packetId, arguments, response)
	srvCmd := NewSrvCmd(command, STATE_RECORDED, SOURCE_SRV)
	s.cmdDb.add(srvCmd)
}

func (s *Server) adminAddCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}

	var command model.CommandBase
	err = json.Unmarshal(reqBody, &command)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  reqBody,
			"error": err,
		}).Info("Error add command")
		return
	}
	srvCmd := NewSrvCmd(command, STATE_RECORDED, SOURCE_SRV)
	s.cmdDb.add(srvCmd)
}
