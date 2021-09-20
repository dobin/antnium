package wingman

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

type Wingman struct {
	conn net.Conn
}

func MakeWingman() Wingman {
	executor := Wingman{}
	return executor
}

func (e *Wingman) StartWingman(destination string) {
	if destination == "" {
		destination = "localhost:50000"
	}
	fmt.Println("Wingman: connect to: " + destination)

	conn, err := net.Dial("tcp", destination)
	if err != nil {
		log.Error("Could not connect: " + err.Error())
		return
	}
	log.Info("Wingman: Connected")

	// Send initial line
	ex, err := os.Executable()
	if err != nil {
		log.Error("Wingman: Error: " + err.Error())
		return
	}
	pid := strconv.Itoa(os.Getpid())
	line := ex + ":" + pid + "\n"
	_, err = conn.Write([]byte(line))
	if err != nil {
		log.Error("Wingman: Error")
		return
	}
	// no answer required
	e.conn = conn

	e.Loop()
}

func (e *Wingman) Shutdown() {
	if e.conn != nil {
		e.conn.Close()
	}
}

func (e *Wingman) Loop() {
	executor := executor.MakeExecutor()

	for {
		// Read
		jsonStr, err := bufio.NewReader(e.conn).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Wingman: Could not read: " + err.Error())
			break
		}
		packet, err := DecodePacket(jsonStr)
		if err != nil {
			log.Error("Wingman: Error: ", err.Error())
			continue
		}

		// Execute
		packet, err = executor.Execute(packet)
		if err != nil {
			log.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Info("Wingman: Error executing packet")

			// TODO ERR
		}

		// Answer: Go to JSON
		packetEncoded, err := EncodePacket(packet)
		if err != nil {
			log.Error("Wingman: Error: ", err.Error())
		}
		n, err := e.conn.Write(packetEncoded)
		if err != nil {
			log.Error("Wingman: Error")

			// TODO ERR
		}
		e.conn.Write([]byte("\n"))
		fmt.Printf("Wingman: Written: %d bytes", n)
	}
}
