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

func (f *FrontendRest) adminListPackets(rw http.ResponseWriter, r *http.Request) {
	packetInfos := f.middleware.FrontendAllPacket()
	json, err := json.Marshal(packetInfos)
	if err != nil {
		e := "Could not JSON marshal"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminListPacketsClientId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientId := vars["clientId"]

	if clientId == "" {
		e := "No client ID given"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	packetInfos := f.middleware.FrontendGetPacketById(clientId)
	json, err := json.Marshal(packetInfos)
	if err != nil {
		e := "Could not JSON marshal"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := f.middleware.FrontendAllClients()
	json, err := json.Marshal(hostList)
	if err != nil {
		e := "Could not JSON marshal"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminUploadFile(rw http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("fileKey")
	if err != nil {
		e := "Error Getting File: "
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = f.middleware.AdminUploadFile(header.Filename, file)
	if err != nil {
		e := fmt.Sprintf("Could not upload file: %s", err.Error())
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, "{ }\n")
}

func (f *FrontendRest) adminAddPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := "Invalid body"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	var packet model.Packet
	err = json.Unmarshal(reqBody, &packet)
	if err != nil {
		e := "Could not JSON marshal"
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}

	common.LogPacket("Add Packet", packet)

	if packet.ClientId == "" || packet.PacketId == "" || packet.PacketType == "" {
		e := fmt.Sprintf("Missing data in packet: %v", packet)
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}

	err = f.middleware.FrontendAddNewPacket(&packet, user)
	if err != nil {
		e := fmt.Sprintf("FrontendAddPacket error: %s", err.Error())
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
}

func (f *FrontendRest) adminGetCampaign(rw http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(f.campaign)
	if err != nil {
		e := fmt.Sprintf("Could not JSON marshal")
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminGetUploads(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./upload")
	if err != nil {
		e := fmt.Sprintf("Could not list directory ./upload: %s", err.Error())
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		e := fmt.Sprintf("Could not JSON marshal")
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminGetStatics(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./static")
	if err != nil {
		e := fmt.Sprintf("Could not list directory ./static: %s", err.Error())
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		e := fmt.Sprintf("Could not JSON marshal")
		log.Error(e)
		http.Error(rw, e, http.StatusBadRequest)
		return
	}
	fmt.Fprint(rw, string(json))
}
