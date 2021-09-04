package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	serverManager *ServerManager
	httpServer    *HttpServer
}

func NewServer(srvAddr string) Server {
	serverManager := NewServerManager(srvAddr)
	httpServer := MakeHttpServer(&serverManager)

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())

	w := Server{
		&serverManager,
		&httpServer,
	}

	return w
}

func (s *Server) Serve() {
	s.httpServer.Serve()
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
		s.serverManager.packetDb.Set(packetInfos)
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
		s.serverManager.clientInfoDb.Set(clients)
		fmt.Printf("Loaded %d clients from %s\n", len(clients), dbClients)
	}

	return nil
}

func (s *Server) DumpDbPackets() error {
	log.Debug("DB Dump: Packets")
	packets := s.serverManager.packetDb.getAll()
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
	clients := s.serverManager.clientInfoDb.getAll()
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
		packets := s.serverManager.packetDb.getAll()
		packetBytes, err := json.Marshal(packets)
		if err != nil {
			log.Errorf("could not marshal config json: %v", err)
		}
		if len(packetBytes) != lastPacketsSize {
			s.DumpDbPackets() // ignore err
			lastPacketsSize = len(packetBytes)
		}

		// Clients
		clients := s.serverManager.clientInfoDb.getAll()
		if len(clients) != lastClientsLen {
			s.DumpDbClients() // ignore err
			lastClientsLen = len(clients)
		}

		time.Sleep(dbDumpInterval)
	}
}
