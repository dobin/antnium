package server

import (
	"time"

	"github.com/dobin/antnium/pkg/model"
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

func (db *ClientInfoDb) updateFor(computerId string, ip string) {
	if _, ok := db.clientInfoDb[computerId]; !ok {
		// Init, without ping (misses a lot of data)
		db.clientInfoDb[computerId] = &ClientInfo{
			computerId,
			time.Now(),
			time.Now(),
			ip,

			"",
			nil,
			"",
			nil,
			"",
		}
	} else {
		// Update
		db.clientInfoDb[computerId].LastSeen = time.Now()
		db.clientInfoDb[computerId].LastIp = ip
	}
}

func (db *ClientInfoDb) updateFromPing(computerId, ip string, response model.PacketResponse) {
	if _, ok := db.clientInfoDb[computerId]; !ok {
		// Init
		db.clientInfoDb[computerId] = &ClientInfo{
			computerId,
			time.Now(),
			time.Now(),
			ip,

			"",
			nil,
			"",
			nil,
			"",
		}
	}

	// Add all relevant data from packet
	hostname, _ := response["hostname"]
	localIps := model.ResponseToArray("localIp", response)
	arch := response["arch"]
	processes := model.ResponseToArray("processes", response)
	InsecureTls := response["InsecureTls"]

	db.clientInfoDb[computerId].Hostname = hostname
	db.clientInfoDb[computerId].LocalIps = localIps
	db.clientInfoDb[computerId].Arch = arch
	db.clientInfoDb[computerId].Processes = processes
	db.clientInfoDb[computerId].InsecureTls = InsecureTls
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
