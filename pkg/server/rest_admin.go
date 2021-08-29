package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *Server) adminListPackets(rw http.ResponseWriter, r *http.Request) {
	packetInfos := s.packetDb.getAll()
	json, err := json.Marshal(packetInfos)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminListPacketsComputerId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	var filteredPackets []PacketInfo = make([]PacketInfo, 0)
	packetInfos := s.packetDb.getAll()
	for _, packetInfo := range packetInfos {
		if packetInfo.Packet.ComputerId == computerId {
			filteredPackets = append(filteredPackets, packetInfo)
		}
	}

	json, err := json.Marshal(filteredPackets)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := s.clientInfoDb.getAsList()
	json, err := json.Marshal(hostList)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminAddTestPacket(rw http.ResponseWriter, r *http.Request) {
	arguments := make(model.PacketArgument)
	//arguments["executable"] = "packet"
	//arguments["arg1"] = "/C"
	//arguments["arg2"] = "whoami"

	//arguments["remoteurl"] = "http://127.0.0.1:4444/psexec.txt"
	//arguments["destination"] = "psexec.txt"

	packetId := s.getRandomPacketId()

	arguments["remoteurl"] = "http://127.0.0.1:4444/upload/" + packetId
	arguments["source"] = "README.md"

	response := make(model.PacketResponse)
	packet := model.NewPacket("fileupload", "0", packetId, arguments, response)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	s.packetDb.add(packetInfo)

	s.adminWebSocket.broadcastPacket(packetInfo)
}

func (s *Server) adminAddPacket(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}

	var packet model.Packet
	err = json.Unmarshal(reqBody, &packet)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  reqBody,
			"error": err,
		}).Info("Error add packet")
		return
	}

	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
	}).Info("Add Packet ")

	if packet.ComputerId == "" || packet.PacketId == "" || packet.PacketType == "" {
		log.Errorf("Missing data in packet: %v", packet)
		return
	}

	packetInfo := NewPacketInfo(packet, STATE_RECORDED)

	// Add to DB and get updated one
	packetInfo = s.packetDb.add(packetInfo)

	// Notify UI immediately (for initial STATE_RECORDED)
	s.adminWebSocket.broadcastPacket(packetInfo)

	// Send it through the pipeline
	s.AddNewPacket(packetInfo)
}

func (s *Server) adminGetCampaign(rw http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(s.Campaign)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminGetUploads(rw http.ResponseWriter, r *http.Request) {
	dirList, err := model.ListDirectory("./upload")
	if err != nil {
		log.Error("Could not: ", err)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		log.Error("Could not JSON marshal", err)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminGetStatics(rw http.ResponseWriter, r *http.Request) {
	dirList, err := model.ListDirectory("./static")
	if err != nil {
		log.Error("Could not: ", err)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		log.Error("Could not JSON marshal", err)
		return
	}
	fmt.Fprint(rw, string(json))
}
