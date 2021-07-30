package client

import (
	"bufio"
	"net"

	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamLocaltcp struct {
	channel chan model.Packet

	packetExecutor executor.PacketExecutor
	conn           net.Conn
}

func MakeDownstreamLocaltcp() DownstreamLocaltcp {
	u := DownstreamLocaltcp{
		make(chan model.Packet),
		executor.MakePacketExecutor(),
		nil,
	}
	return u
}

func (d *DownstreamLocaltcp) start() {
	go d.startServer()

	for {
		packet := <-d.channel // Wait for new packet for this downstream

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
		d.channel <- packet
	}
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
	}
}
