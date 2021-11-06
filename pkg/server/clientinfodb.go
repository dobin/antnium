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

func (db *ClientInfoDb) updateFor(clientId string, ip string, connectorType string) {
	if _, ok := db.clients[clientId]; !ok {
		log.Infof("New client %s: %s via %s", ip, clientId, connectorType)
		// Init, without ping (misses a lot of data)
		db.clients[clientId] = &ClientInfo{
			ClientId:  clientId,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			LastIp:    ip,

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
		db.clients[clientId].LastSeen = time.Now()
		db.clients[clientId].LastIp = ip
		db.clients[clientId].ConnectorType = connectorType
	}
}

func (db *ClientInfoDb) updateFromClientinfo(clientId, ip string, connectorType string, response model.PacketResponse) {
	if _, ok := db.clients[clientId]; !ok {
		// Init
		db.clients[clientId] = &ClientInfo{
			ClientId:  clientId,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			LastIp:    ip,
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

	db.clients[clientId].Hostname = hostname
	db.clients[clientId].LocalIps = localIps
	db.clients[clientId].Arch = arch
	db.clients[clientId].Processes = processes
	db.clients[clientId].IsAdmin = isAdmin
	db.clients[clientId].IsElevated = isElevated
	db.clients[clientId].WorkingDir = WorkingDir
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
