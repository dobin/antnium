package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

/*
	// Clients connected via websocket do not send regular ping packets (that's the idea of it)
	// Sadly this makes LastSeen useless - but the user wants to know if the client is still connected.
	// Here we regularly check the clients connected to ClientWebsocket, and update their LastSeen
	// Lifetime: App
	go func() {
		clientInfoDb2 := &clientInfoDb
		for {
			time.Sleep(10 * time.Second)
			c := clientWebsocket.clients
			for computerId, conn := range c {
				if conn == nil {
					continue
				}
				clientInfoDb2.updateFor(computerId, conn.RemoteAddr().String())
			}

			// Todo: When to quit?
		}
	}()
*/

type Server struct {
	srvaddr          string
	Campaign         *campaign.Campaign
	connectorManager *ConnectorManager
	frontendManager  *FrontendManager
	middleware       *Middleware
	wsUpgrader       websocket.Upgrader
}

func NewServer(srvAddr string) Server {
	campaign := campaign.MakeCampaign()
	middleware := MakeMiddleware()
	connectorManager := MakeConnectorManager(&campaign, &middleware)
	frontendManager := MakeFrontendManager(&campaign, &middleware)
	middleware.SetTODO(&connectorManager, &frontendManager)

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())

	w := Server{
		srvAddr,
		&campaign,
		&connectorManager,
		&frontendManager,
		&middleware,

		websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}

	return w
}

func (s *Server) DbLoad() error {
	// Packets
	dbPackets := "db.packets.json"
	if _, err := os.Stat(dbPackets); !os.IsNotExist(err) {
		packetsBytes, err := ioutil.ReadFile(dbPackets)
		if err != nil {
			return fmt.Errorf("Read file error: %s", err.Error())
		}
		var packetInfos []PacketInfo
		err = json.Unmarshal(packetsBytes, &packetInfos)
		if err != nil {
			return fmt.Errorf("Read file decode error: %s", err.Error())
		}
		s.middleware.packetDb.Set(packetInfos)
		fmt.Printf("Loaded %d packets from %s\n", len(packetInfos), dbPackets)
	}

	// Clients
	dbClients := "db.clients.json"
	if _, err := os.Stat(dbClients); !os.IsNotExist(err) {
		clientsBytes, err := ioutil.ReadFile(dbClients)
		if err != nil {
			return fmt.Errorf("Read file error: %s", err.Error())
		}
		var clients map[string]*ClientInfo
		err = json.Unmarshal(clientsBytes, &clients)
		if err != nil {
			return fmt.Errorf("Read file decode error: %s", err.Error())
		}
		s.middleware.clientInfoDb.Set(clients)
		fmt.Printf("Loaded %d clients from %s\n", len(clients), dbClients)
	}

	return nil
}

func (s *Server) DumpDbPackets() error {
	log.Debug("DB Dump: Packets")
	packets := s.middleware.packetDb.getAll()
	packetBytes, err := json.Marshal(packets)
	if err != nil {
		log.Errorf("could not marshal config json: %v", err)
		return err
	}

	err = ioutil.WriteFile("db.packets.json", packetBytes, 0644)
	if err != nil {
		log.Errorf("could not marshal config json: %v", err)
		return err
	}

	return nil
}

func (s *Server) DumpDbClients() error {
	log.Debug("DB Dump: Clients")
	clients := s.middleware.clientInfoDb.getAll()
	clientsBytes, err := json.Marshal(clients)
	if err != nil {
		log.Errorf("could not marshal config json: %v", err)
		return err
	}
	err = ioutil.WriteFile("db.clients.json", clientsBytes, 0644)
	if err != nil {
		log.Errorf("could not marshal config json: %v", err)
		return err
	}

	return nil
}

// PeriodicDbDump is a Thread which regularly dumps the memory content to file system, lifetime:app
func (s *Server) PeriodicDbDump() {
	dbDumpInterval := 1 * time.Minute

	lastPacketsSize := 0 // Can't take len, as packets could be updated
	lastClientsLen := 0  // len of array. at least we get all clients
	for {
		// Packets
		packets := s.middleware.packetDb.getAll()
		packetBytes, err := json.Marshal(packets)
		if err != nil {
			log.Errorf("could not marshal config json: %v", err)
		}
		if len(packetBytes) != lastPacketsSize {
			s.DumpDbPackets() // ignore err
			lastPacketsSize = len(packetBytes)
		}

		// Clients
		clients := s.middleware.clientInfoDb.getAll()
		if len(clients) != lastClientsLen {
			s.DumpDbClients() // ignore err
			lastClientsLen = len(clients)
		}

		time.Sleep(dbDumpInterval)
	}
}
