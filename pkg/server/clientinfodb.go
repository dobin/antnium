package server

import (
	"time"

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

func (db *ClientInfoDb) updateFor(computerId string, ip string) {
	if _, ok := db.clientInfoDb[computerId]; !ok {
		db.clientInfoDb[computerId] = &ClientInfo{
			computerId,
			time.Now(),
			time.Now(),
			ip,
			"",
			nil,
		}
	} else {
		db.clientInfoDb[computerId].LastSeen = time.Now()
		db.clientInfoDb[computerId].LastIp = ip
	}
}

func (db *ClientInfoDb) updateMore(computerId, hostname string, localIps []string) {
	if _, ok := db.clientInfoDb[computerId]; ok {
		db.clientInfoDb[computerId].Hostname = hostname
		db.clientInfoDb[computerId].LocalIps = localIps
	} else {
		log.Error("Client not found in clientdb")
	}
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
