package common

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/aus/proxyplease"
	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

func GetRandomPacketId() string {
	buf := make([]byte, 8)
	_, err := crypto_rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}

	data := binary.BigEndian.Uint64(buf)
	return strconv.FormatUint(data, 16)
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
		"1_clientId":     packet.ClientId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Info(s)
}

func LogPacketDebug(s string, packet model.Packet) {
	log.WithFields(log.Fields{
		"1_clientId":     packet.ClientId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Debug(s)
}

// FreePort asks the kernel for a free open port that is ready to use.
func FreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Error("ResolveTCPAddr")
		port := 50000 + rand.Intn(9999)
		return strconv.Itoa(port), nil
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error("ListenTCP test")
		port := 50000 + rand.Intn(9999)
		return strconv.Itoa(port), nil
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	return strconv.Itoa(port), nil
}

func NewDialContext(campaign *campaign.Campaign) (proxyplease.DialContext, error) {
	var dialContext proxyplease.DialContext

	if campaign.DisableProxy {
		log.Info("Disabled proxy, use direct")
		dialContext = (&net.Dialer{
			KeepAlive: 5 * time.Second,
		}).DialContext

		return dialContext, nil
	}

	// Automatic proxy configuration
	proxyUrl, ok := campaign.GetProxy()
	if ok {
		// Manual proxy configuration
		parsedUrl, err := url.Parse(proxyUrl)
		if err != nil {
			//return fmt.Errorf("could not parse ProxyUrl %s: %s", proxyUrl, err.Error())
			return nil, fmt.Errorf("could not parse ProxyUrl %s: %s", proxyUrl, err.Error())
		}
		dialContext = proxyplease.NewDialContext(proxyplease.Proxy{URL: parsedUrl})
	} else {
		dialContext = proxyplease.NewDialContext(proxyplease.Proxy{})
	}
	return dialContext, nil
}
