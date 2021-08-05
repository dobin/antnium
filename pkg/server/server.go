package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	srvaddr        string
	campaign       model.Campaign
	coder          model.Coder
	packetDb       PacketDb
	clientInfoDb   ClientInfoDb
	adminWebSocket AdminWebSocket
}

func NewServer(srvAddr string) Server {
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(&campaign)
	packetDb := MakePacketDb()
	clientInfoDb := MakeClientInfoDb()
	adminWebsocket := MakeAdminWebSocket(campaign.AdminApiKey)

	w := Server{
		srvAddr,
		campaign,
		coder,
		packetDb,
		clientInfoDb,
		adminWebsocket,
	}

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())

	return w
}

func (s *Server) PeriodicDbDump() {
	dbDumpInterval := 1 * time.Minute

	lastPacketsSize := 0 // Can't take len, as packets could be updated
	lastClientsLen := 0  // len of array. at least we get all clients
	for {
		// Packets
		packets := s.packetDb.getAll()
		packetBytes, err := json.Marshal(packets)
		if err != nil {
			log.Errorf("could not marshal config json: %v", err)
		}
		if len(packetBytes) != lastPacketsSize {
			fmt.Println("Write1")
			err = ioutil.WriteFile("db.packets.json", packetBytes, 0644)
			if err != nil {
				log.Errorf("could not marshal config json: %v", err)
			}
			lastPacketsSize = len(packetBytes)
		}

		// Clients
		clients := s.clientInfoDb.getAll()
		if len(clients) != lastClientsLen {
			clientsBytes, err := json.Marshal(clients)
			if err != nil {
				log.Errorf("could not marshal config json: %v", err)
			}
			fmt.Println("Write2")
			err = ioutil.WriteFile("db.clients.json", clientsBytes, 0644)
			if err != nil {
				log.Errorf("could not marshal config json: %v", err)
			}
			lastClientsLen = len(clients)
		}

		time.Sleep(dbDumpInterval)
	}

}
