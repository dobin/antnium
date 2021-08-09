package client

import (
	"net"
	"os"
	"runtime"

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

	InsecureTls bool // If we should accept invalid TLS certs
}

func MakeClientConfig() ClientConfig {
	// Hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Error("Hostname failed")
		hostname = "unknown"
	}

	// Local interfaces
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

	// Arch
	arch := "unknown"
	if runtime.GOOS == "windows" {
		arch = "windows"
	} else if runtime.GOOS == "linux" {
		arch = "linux"
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
		xid.New().String(),
		hostname,
		localIps,
		arch,
		processList,
		true,
	}
	return db
}
