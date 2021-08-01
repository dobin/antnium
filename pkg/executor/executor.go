package executor

import (
	"bufio"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

type Executor struct {
}

func MakeExecutor() Executor {
	executor := Executor{}
	return executor
}

func (e *Executor) StartClient() {
	destination := "localhost:50000"
	fmt.Println("Executor connect to: " + destination)

	conn, err := net.Dial("tcp", destination)
	if err != nil {
		log.Error("Could not connect: " + err.Error())
	}
	log.Info("Executor connected")

	e.Loop(conn)
}

func (e *Executor) Loop(conn net.Conn) {
	packetExecutor := MakePacketExecutor()

	for {
		// Read
		jsonStr, err := bufio.NewReader(conn).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Could not read: " + err.Error())
			break
		}

		fmt.Println("Jsonstr: " + jsonStr)
		packet, err := DecodePacket(jsonStr)
		if err != nil {
			log.Error("Error: ", err.Error())
		}

		// Execute
		err = packetExecutor.Execute(&packet)
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

		n, err := conn.Write(packetEncoded)
		if err != nil {
			log.Error("Error")

			// TODO ERR
		}
		conn.Write([]byte("\n"))
		fmt.Printf("Written: %d bytes", n)

	}
}
