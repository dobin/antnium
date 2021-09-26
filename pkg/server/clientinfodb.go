package server

import (
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type ClientInfoMap map[string]*ClientInfo

type ClientInfoDb struct {
	// Needs to be a pointer to ClientInfo so we can change its values
	clients ClientInfoMap
}

func MakeClientInfoDb() ClientInfoDb {
	db := ClientInfoDb{
		make(ClientInfoMap),
	}
	return db
}

func (db *ClientInfoDb) updateFor(computerId string, ip string, connectorType string) {
	if _, ok := db.clients[computerId]; !ok {
		log.Infof("New client %s: %s via %s", ip, computerId, connectorType)
		// Init, without ping (misses a lot of data)
		db.clients[computerId] = &ClientInfo{
			ComputerId: computerId,
			FirstSeen:  time.Now(),
			LastSeen:   time.Now(),
			LastIp:     ip,

			Hostname:      "",
			LocalIps:      nil,
			Arch:          "",
			Processes:     nil,
			IsAdmin:       "",
			IsElevated:    "",
			ConnectorType: connectorType,
		}
	} else {
		// Update
		db.clients[computerId].LastSeen = time.Now()
		db.clients[computerId].LastIp = ip
		db.clients[computerId].ConnectorType = connectorType
	}
}

func (db *ClientInfoDb) updateFromClientinfo(computerId, ip string, connectorType string, response model.PacketResponse) {
	if _, ok := db.clients[computerId]; !ok {
		// Init
		db.clients[computerId] = &ClientInfo{
			ComputerId: computerId,
			FirstSeen:  time.Now(),
			LastSeen:   time.Now(),
			LastIp:     ip,
		}
	}

	// Add all relevant data from packet
	hostname, _ := response["hostname"]
	if hostname == "" {
		log.Warn("ClientInfoDb: Empty ping")
		return
	}
	localIps := model.ResponseToArray("localIp", response)
	arch := response["arch"]
	isAdmin := response["isAdmin"]
	isElevated := response["isElevated"]
	processes := model.ResponseToArray("processes", response)
	WorkingDir := response["WorkingDir"]

	db.clients[computerId].Hostname = hostname
	db.clients[computerId].LocalIps = localIps
	db.clients[computerId].Arch = arch
	db.clients[computerId].Processes = processes
	db.clients[computerId].IsAdmin = isAdmin
	db.clients[computerId].IsElevated = isElevated
	db.clients[computerId].WorkingDir = WorkingDir
}

func (db *ClientInfoDb) AllAsList() []ClientInfo {
	v := make([]ClientInfo, 0, len(db.clients))
	for _, value := range db.clients {
		v = append(v, *value)
	}
	return v
}

func (db *ClientInfoDb) All() ClientInfoMap {
	return db.clients
}

func (db *ClientInfoDb) Set(clients ClientInfoMap) {
	db.clients = clients
}
