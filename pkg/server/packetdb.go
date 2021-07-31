package server

import (
	"fmt"
	"time"

	"github.com/dobin/antnium/pkg/model"
)

type PacketDb struct {
	packetInfo []PacketInfo
}

func MakePacketDb() PacketDb {
	db := PacketDb{
		make([]PacketInfo, 0),
	}
	return db
}

func (db *PacketDb) add(packetInfo PacketInfo) PacketInfo {
	packetInfo.TimeRecorded = time.Now()
	db.packetInfo = append(db.packetInfo, packetInfo)
	return packetInfo
}

func (db *PacketDb) getAll() []PacketInfo {
	return db.packetInfo
}

func (db *PacketDb) ByPacketId(packetId string) (PacketInfo, error) {
	for _, packetInfo := range db.packetInfo {
		if packetInfo.Packet.PacketId == packetId {
			return packetInfo, nil
		}
	}

	return PacketInfo{}, fmt.Errorf("Nothing found")
}

func (db *PacketDb) getPacketFor(computerId string) (*PacketInfo, error) {
	for i, packetInfo := range db.packetInfo {
		if packetInfo.State != STATE_RECORDED {
			continue
		}
		packetInfoComputerId := packetInfo.Packet.ComputerId
		if packetInfoComputerId == "0" || packetInfoComputerId == computerId {
			db.packetInfo[i].State = STATE_SENT // FIXME
			db.packetInfo[i].TimeSent = time.Now()
			return &db.packetInfo[i], nil
		}
	}

	return &PacketInfo{}, fmt.Errorf("Nothing found")
}

func (db *PacketDb) update(packet model.Packet) PacketInfo {
	// Update existing
	for i, packetInfo := range db.packetInfo {
		if packetInfo.Packet.PacketId == packet.PacketId {
			db.packetInfo[i].State = STATE_ANSWERED
			db.packetInfo[i].TimeAnswered = time.Now()
			db.packetInfo[i].Packet.Response = packet.Response
			db.packetInfo[i].Packet.ComputerId = packet.ComputerId
			return db.packetInfo[i]
		}
	}

	// Add new (client initiated)
	fakePacketInfo := NewPacketInfo(packet, STATE_CLIENT)
	db.add(fakePacketInfo)
	return fakePacketInfo
}
