package client

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamLocaltcp struct {
	packetExecutor executor.PacketExecutor
	conns          []net.Conn
}

func MakeDownstreamLocaltcp() DownstreamLocaltcp {
	u := DownstreamLocaltcp{
		executor.MakePacketExecutor(),
		nil,
	}
	return u
}

func (d *DownstreamLocaltcp) getList() []string {
	ret := make([]string, 0)
	for i, _ := range d.conns {
		name := "net#" + strconv.Itoa(i)
		ret = append(ret, name)
	}
	return ret
}

func (d *DownstreamLocaltcp) do(packet model.Packet) (model.Packet, error) {
	if len(d.conns) == 0 {
		log.Error("No downstream clients")
		packet.Response["error"] = "No downstream clients found"
		return packet, fmt.Errorf("No downstream clients found")
	}

	split := strings.Split(packet.DownstreamId, "#")
	if len(split) != 2 {
		return packet, fmt.Errorf("Wrong format")
	}
	downstreamIdStr := split[1]
	downstreamId, err := strconv.Atoi(downstreamIdStr)
	if err != nil {
		log.Error(err)
		return packet, fmt.Errorf("Invalid number")
	}
	if downstreamId < 0 || downstreamId >= len(d.conns) {
		log.Error("DownstreamId does not exist")
		packet.Response["error"] = "DownstreamId does not exist"
		return packet, fmt.Errorf("DownstreamId does not exist")
	}

	// If conn down...
	return d.doConn(d.conns[downstreamId], packet)
}

func (d *DownstreamLocaltcp) doConn(conn net.Conn, packet model.Packet) (model.Packet, error) {
	// Send it to the downstream executor
	packetEncoded, err := executor.EncodePacket(packet)
	if err != nil {
		log.Error("Error: ", err.Error())
		packet.Response["error"] = err.Error()
		return packet, err
	}
	_, err = conn.Write(packetEncoded)
	if err != nil {
		log.Error("Could not write: " + err.Error())
		packet.Response["error"] = err.Error()
		return packet, err
	}
	conn.Write([]byte("\n"))

	// Wait for answer
	jsonStr, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Error("Could not read: " + err.Error())
		packet.Response["error"] = err.Error()
		return packet, err
	}
	packet, err = executor.DecodePacket(jsonStr)
	if err != nil {
		log.Error("Error: ", err.Error())
		packet.Response["error"] = err.Error()
		return packet, err
	}

	return packet, nil
}

func (d *DownstreamLocaltcp) startServer(c chan []string) {
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
			continue
		}
		d.conns = append(d.conns, conn)

		// Notify about new downstream
		c <- d.getList()
	}
}
