package client

import (
	"net"
	"os"

	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

type ClientConfig struct {
	ComputerId string //
	Hostname   string
	LocalIps   []string
}

func MakeClientConfig() ClientConfig {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error("Hostname failed")
		hostname = "unknown"
	}

	localIps := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error("Local interfaces: Could not get interfaces")
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Error("Local interfaces: Could not handle IP address")
		}
		// handle err
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

	db := ClientConfig{
		xid.New().String(),
		hostname,
		localIps,
	}
	return db
}
