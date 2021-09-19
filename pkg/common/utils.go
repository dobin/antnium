package common

import (
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

func GetRandomPacketId() string {
	return strconv.FormatUint(rand.Uint64(), 16)
}

type DirEntry struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Mode     string    `json:"mode"`
	Modified time.Time `json:"modified"`
	IsDir    bool      `json:"isDir"`
}

func ListDirectory(path string) ([]DirEntry, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dirList := make([]DirEntry, 0)
	for _, file := range files {
		dl := DirEntry{
			file.Name(),
			file.Size(),
			"", // Mode()
			file.ModTime(),
			file.IsDir(),
		}
		dirList = append(dirList, dl)
	}

	return dirList, err
}

func LogPacket(s string, packet model.Packet) {
	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Info(s)
}

func LogPacketDebug(s string, packet model.Packet) {
	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Debug(s)
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	return strconv.Itoa(port), nil
}
