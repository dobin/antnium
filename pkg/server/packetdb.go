package server

import (
	"fmt"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type PacketDb struct {
	packets []*PacketInfo
}

func MakePacketDb() PacketDb {
	db := PacketDb{
		make([]*PacketInfo, 0, 512),
	}
	return db
}

func (db *PacketDb) add(packetInfo *PacketInfo) {
	packetInfo.TimeRecorded = time.Now()

	db.packets = append(db.packets, packetInfo)
}

func (db *PacketDb) All() []*PacketInfo {
	return db.packets
}

func (db *PacketDb) Set(packetInfos []*PacketInfo) {
	db.packets = packetInfos
}

func (db *PacketDb) ByPacketId(packetId string) (*PacketInfo, bool) {
	for _, packetInfo := range db.packets {
		if packetInfo.Packet.PacketId == packetId {
			return packetInfo, true
		}
	}

	return nil, false
}

func (db *PacketDb) getAllUnsentPacketsToClient(clientId string) []PacketInfo {
	packetInfos := make([]PacketInfo, 0)
	for _, packetInfo := range db.packets {
		if packetInfo.State != STATE_RECORDED {
			continue
		}
		if packetInfo.Packet.ClientId != clientId {
			continue
		}
		packetInfos = append(packetInfos, *packetInfo)
	}

	return packetInfos
}

func (db *PacketDb) getUnsentPacketForClient(clientId string) (*PacketInfo, error) {
	for _, packetInfo := range db.packets {
		if packetInfo.State != STATE_RECORDED {
			continue
		}
		if packetInfo.Packet.ClientId == clientId {
			return packetInfo, nil
		}
	}

	return nil, fmt.Errorf("no packets in state STATE_RECORDED for host %s found", clientId)
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
		log.Warnf("PacketDb: wrong packet source state for packetDb.Update(), expect STATE_SENT, got %d", packetInfo.State)
	}
	if packet.Response == nil {
		// Can this be?
		log.Warnf("Packet Response nil")
		packet.Response = make(model.PacketResponse)
	}
	packetInfo.State = STATE_ANSWERED
	packetInfo.TimeAnswered = time.Now()
	packetInfo.Packet.Response = packet.Response

	return packetInfo
}

func (db *PacketDb) addFromFrontend(packet *model.Packet, user string) (*PacketInfo, error) {
	_, ok := db.ByPacketId(packet.PacketId)
	if ok {
		return nil, fmt.Errorf("PacketId %s already exists in DB. Wont handle it.", packet.PacketId)
	}

	// Add new (always client initiated for now)
	packetInfo := NewPacketInfo(*packet, STATE_RECORDED)
	packetInfo.TimeRecorded = time.Now()
	packetInfo.User = user

	db.add(&packetInfo)
	return &packetInfo, nil
}

func (db *PacketDb) sentToClient(packetId string, remoteAddr string) (*PacketInfo, error) {
	packetInfo, ok := db.ByPacketId(packetId)
	if !ok {
		return nil, fmt.Errorf("PacketDb: Packet with PacketId %s does not exist", packetId)
	}

	if packetInfo.State != STATE_RECORDED {
		return nil, fmt.Errorf("source packet not STATE_RECORDED but %d", packetInfo.State)
	}

	packetInfo.State = STATE_SENT
	packetInfo.TimeSent = time.Now()

	return packetInfo, nil
}
