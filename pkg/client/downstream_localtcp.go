package client

import (
	"bufio"
	"net"

	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamLocaltcp struct {
	packetExecutor executor.PacketExecutor
	conn           net.Conn
}

func MakeDownstreamLocaltcp() DownstreamLocaltcp {
	u := DownstreamLocaltcp{
		executor.MakePacketExecutor(),
		nil,
	}
	return u
}

func (d *DownstreamLocaltcp) do(packet model.Packet) (model.Packet, error) {
	// Send it to the downstream executor
	packetEncoded, err := executor.EncodePacket(packet)
	if err != nil {
		log.Error("Error: ", err.Error())
	}
	d.conn.Write(packetEncoded)
	d.conn.Write([]byte("\n"))

	// Wait for answer
	jsonStr, err := bufio.NewReader(d.conn).ReadString('\n')
	if err != nil {
		log.Error("Could not read: " + err.Error())
	}
	packet, err = executor.DecodePacket(jsonStr)
	if err != nil {
		log.Error("Error: ", err.Error())
	}

	// Always send response, as it is synchronized
	if err != nil {
		packet.Response["error"] = err.Error()
	}

	return packet, err
}

func (d *DownstreamLocaltcp) startServer() {
	log.Info("Start Downstream: LocalTcp")
	ln, err := net.Listen("tcp", "127.0.0.1:50000")
	if err != nil {
		log.Error("Error: " + err.Error())
		// TODO: Handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Error: " + err.Error())
			// TODO: Handle error
		}
		d.conn = conn

		// SEND TO SERVER, in DownstreamManager, via Client?
	}
}
