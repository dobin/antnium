package server

import (
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type ClientInfoDb struct {
	// Needs to be a pointer to ClientInfo so we can change its values
	clientInfoDb map[string]*ClientInfo
}

func MakeClientInfoDb() ClientInfoDb {
	db := ClientInfoDb{
		make(map[string]*ClientInfo),
	}
	return db
}

func (db *ClientInfoDb) updateFor(computerId string, ip string, connectorType string) {
	if _, ok := db.clientInfoDb[computerId]; !ok {
		log.Infof("New client %s: %s via %s", ip, computerId, connectorType)
		// Init, without ping (misses a lot of data)
		db.clientInfoDb[computerId] = &ClientInfo{
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
		db.clientInfoDb[computerId].LastSeen = time.Now()
		db.clientInfoDb[computerId].LastIp = ip
		db.clientInfoDb[computerId].ConnectorType = connectorType
	}
}

func (db *ClientInfoDb) updateFromClientinfo(computerId, ip string, connectorType string, response model.PacketResponse) {
	if _, ok := db.clientInfoDb[computerId]; !ok {
		// Init
		db.clientInfoDb[computerId] = &ClientInfo{
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
			ConnectorType: "",
		}
	}

	// Add all relevant data from packet
	hostname, _ := response["hostname"]
	localIps := model.ResponseToArray("localIp", response)
	arch := response["arch"]
	isAdmin := response["isAdmin"]
	isElevated := response["isElevated"]
	processes := model.ResponseToArray("processes", response)

	if hostname == "" {
		log.Warn("ClientInfoDb: Empty ping")
		return
	}

	db.clientInfoDb[computerId].Hostname = hostname
	db.clientInfoDb[computerId].LocalIps = localIps
	db.clientInfoDb[computerId].Arch = arch
	db.clientInfoDb[computerId].Processes = processes
	db.clientInfoDb[computerId].IsAdmin = isAdmin
	db.clientInfoDb[computerId].IsElevated = isElevated
}

func (db *ClientInfoDb) getAsList() []ClientInfo {
	v := make([]ClientInfo, 0, len(db.clientInfoDb))
	for _, value := range db.clientInfoDb {
		v = append(v, *value)
	}
	return v
}

func (db *ClientInfoDb) getAll() map[string]*ClientInfo {
	return db.clientInfoDb
}

func (db *ClientInfoDb) Set(clients map[string]*ClientInfo) {
	db.clientInfoDb = clients
}
