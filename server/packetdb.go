package server

import (
	"fmt"
	"time"

	"github.com/dobin/antnium/model"
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
		if packetInfo.Command.PacketId == packetId {
			return packetInfo, nil
		}
	}

	return PacketInfo{}, fmt.Errorf("Nothing found")
}

func (db *PacketDb) getCommandFor(computerId string) (*PacketInfo, error) {
	for i, packetInfo := range db.packetInfo {
		if packetInfo.State != STATE_RECORDED {
			continue
		}
		packetInfoComputerId := packetInfo.Command.ComputerId
		if packetInfoComputerId == "0" || packetInfoComputerId == computerId {
			db.packetInfo[i].State = STATE_SENT // FIXME
			db.packetInfo[i].TimeSent = time.Now()
			return &db.packetInfo[i], nil
		}
	}

	return &PacketInfo{}, fmt.Errorf("Nothing found")
}

func (db *PacketDb) update(command model.Packet) (PacketInfo, error) {
	for i, packetInfo := range db.packetInfo {
		if packetInfo.Command.PacketId == command.PacketId {
			db.packetInfo[i].State = STATE_ANSWERED
			db.packetInfo[i].TimeAnswered = time.Now()
			db.packetInfo[i].Command.Response = command.Response
			db.packetInfo[i].Command.ComputerId = command.ComputerId
			return db.packetInfo[i], nil
		}
	}

	return PacketInfo{}, fmt.Errorf("command not found")
}
