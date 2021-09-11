package server

import (
	"fmt"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type PacketDb struct {
	packetInfo []*PacketInfo
}

func MakePacketDb() PacketDb {
	db := PacketDb{
		make([]*PacketInfo, 0, 512),
	}
	return db
}

func (db *PacketDb) add(packetInfo *PacketInfo) {
	packetInfo.TimeRecorded = time.Now()
	db.packetInfo = append(db.packetInfo, packetInfo)
}

func (db *PacketDb) getAll() []*PacketInfo {
	return db.packetInfo
}

func (db *PacketDb) Set(packetInfos []*PacketInfo) {
	db.packetInfo = packetInfos
}

func (db *PacketDb) ByPacketId(packetId string) (*PacketInfo, bool) {
	for _, packetInfo := range db.packetInfo {
		if packetInfo.Packet.PacketId == packetId {
			return packetInfo, true
		}
	}

	return nil, false
}

func (db *PacketDb) getPacketForClient(computerId string) (*PacketInfo, error) {
	for _, packetInfo := range db.packetInfo {
		if packetInfo.State != STATE_RECORDED {
			continue
		}
		if packetInfo.Packet.ComputerId == computerId {
			return packetInfo, nil
		}
	}

	return nil, fmt.Errorf("Nothing found")
}

func (db *PacketDb) updateFromClient(packet model.Packet) *PacketInfo {
	packetInfo, ok := db.ByPacketId(packet.PacketId)
	if !ok {
		// Add new (always client initiated for now)
		packetInfo := NewPacketInfo(packet, STATE_CLIENT)
		t := time.Now()
		packetInfo.TimeRecorded = t
		packetInfo.TimeAnswered = t
		db.add(&packetInfo)
		return &packetInfo
	}

	if packetInfo.State != STATE_SENT {
		log.Errorf("Wrong packet source state for packetDb.Update(), expect STATE_SENT, got %d", packetInfo.State)
	}
	packetInfo.State = STATE_ANSWERED
	packetInfo.TimeAnswered = time.Now()
	packetInfo.Packet.Response = packet.Response

	return packetInfo
}

func (db *PacketDb) addFromAdmin(packet model.Packet) *PacketInfo {
	// Add new (always client initiated for now)
	packetInfo := NewPacketInfo(packet, STATE_RECORDED)
	packetInfo.TimeRecorded = time.Now()

	db.add(&packetInfo)
	return &packetInfo
}

func (db *PacketDb) sentToClient(packetId string, remoteAddr string) (*PacketInfo, error) {
	packetInfo, ok := db.ByPacketId(packetId)
	if !ok {
		return nil, fmt.Errorf("sentToClient: Packet %s does not exist ", packetId)
	}

	if packetInfo.State != STATE_RECORDED {
		log.Warnf("sentToClient: source packet not STATE_RECORDED: %d", packetInfo.State)
		return nil, fmt.Errorf("sentToClient: source packet not STATE_RECORDED: %d", packetInfo.State)
	}

	packetInfo.ClientIp = remoteAddr
	packetInfo.State = STATE_SENT
	packetInfo.TimeSent = time.Now()

	return packetInfo, nil
}
