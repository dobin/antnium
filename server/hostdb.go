package server

import (
	"time"
)

type HostDb struct {
	// Needs to be a pointer to HostBase so we can change its values
	hostDb map[string]*HostBase
}

func MakeHostDb() HostDb {
	db := HostDb{
		make(map[string]*HostBase),
	}
	return db
}

func (db *HostDb) updateFor(computerId string, ip string) {
	if _, ok := db.hostDb[computerId]; !ok {
		db.hostDb[computerId] = &HostBase{
			computerId,
			time.Now(),
			ip,
		}
	} else {
		db.hostDb[computerId].LastSeen = time.Now()
	}
}

func (db *HostDb) getAsList() []HostBase {
	v := make([]HostBase, 0, len(db.hostDb))
	for _, value := range db.hostDb {
		v = append(v, *value)
	}
	return v
}
