package client

import (
	"bufio"
	"fmt"
	"net"
	"strconv"

	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamInfoTcp struct {
	Name string
	Info string
	conn net.Conn
}
type DownstreamInfoTcpMap map[string]DownstreamInfoTcp

type DownstreamLocaltcp struct {
	listenAddr     string
	packetExecutor executor.PacketExecutor
	downstreams    DownstreamInfoTcpMap
}

func MakeDownstreamLocaltcp(listenAddr string) DownstreamLocaltcp {
	// Default
	if listenAddr == "" {
		listenAddr = "localhost:50000"
	}

	u := DownstreamLocaltcp{
		listenAddr,
		executor.MakePacketExecutor(),
		make(DownstreamInfoTcpMap, 0),
	}
	return u
}

func (d *DownstreamLocaltcp) do(packet model.Packet) (model.Packet, error) {
	downstreamInfo, ok := d.downstreams[packet.DownstreamId]
	if !ok {
		log.Errorf("Did not find downstream: %s in %v", packet.DownstreamId, d.downstreams)
		return model.Packet{}, fmt.Errorf("Did not find: %s", packet.DownstreamId)
	}

	return d.doConn(downstreamInfo.conn, packet)
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

func (d *DownstreamLocaltcp) startServer(downstreamLocaltcpChannel chan struct{}) {
	log.Info("Start Downstream: LocalTcp on " + d.listenAddr)
	ln, err := net.Listen("tcp", d.listenAddr)
	if err != nil {
		log.Error("Error: " + err.Error())
		// TODO: Handle error
	}

	n := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Error: " + err.Error())
			// TODO: Handle error
			continue
		}

		// receive info line first or fail
		infoStr, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Error("Could not read: " + err.Error())
			continue
		}

		// Add it to local datastructure
		name := "net#" + strconv.Itoa(n)
		info := infoStr
		downstreamInfoTcp := DownstreamInfoTcp{
			name,
			info,
			conn,
		}
		d.downstreams[name] = downstreamInfoTcp

		// Notify about new downstream
		downstreamLocaltcpChannel <- struct{}{}

		n += 1
	}
}

func (d *DownstreamLocaltcp) DownstreamList() []DownstreamInfo {
	ret := make([]DownstreamInfo, 0)

	for _, downstreamInfoTcp := range d.downstreams {
		d := DownstreamInfo{
			downstreamInfoTcp.Name,
			downstreamInfoTcp.Info,
		}
		ret = append(ret, d)
	}

	return ret
}
