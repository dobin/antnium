package client

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/dobin/antnium/pkg/downstreamclient"
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
	listenAddr       string               // which TCP port address we listen on
	downstreams      DownstreamInfoTcpMap // all ever accepted connections
	downstreamsMutex *sync.Mutex          // downstreams map updated via startServer() thread
}

func MakeDownstreamLocaltcp(listenAddr string) DownstreamLocaltcp {
	// Default
	if listenAddr == "" {
		listenAddr = "localhost:50000"
	}

	u := DownstreamLocaltcp{
		listenAddr,
		make(DownstreamInfoTcpMap, 0),
		&sync.Mutex{},
	}
	return u
}

func (d *DownstreamLocaltcp) do(packet model.Packet) (model.Packet, error) {
	d.downstreamsMutex.Lock()
	downstreamInfo, ok := d.downstreams[packet.DownstreamId]
	d.downstreamsMutex.Unlock()
	if !ok {
		log.Errorf("Did not find downstream: %s in %v", packet.DownstreamId, d.downstreams)
		return model.Packet{}, fmt.Errorf("Did not find: %s", packet.DownstreamId)
	}

	packet, err := d.doConn(downstreamInfo.conn, packet)
	if err != nil {
		log.Error("Error: ", err.Error())
		// Add error to packet response
		packet.Response["error"] = err.Error()
	}
	return packet, err
}

func (d *DownstreamLocaltcp) doConn(conn net.Conn, packet model.Packet) (model.Packet, error) {
	// Send it to the downstream executor
	packetEncoded, err := downstreamclient.EncodePacket(packet)
	if err != nil {
		return packet, err
	}
	_, err = conn.Write(packetEncoded)
	if err != nil {
		return packet, err
	}
	conn.Write([]byte("\n"))

	// Wait for answer
	jsonStr, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return packet, err
	}
	packet, err = downstreamclient.DecodePacket(jsonStr)
	if err != nil {
		return packet, err
	}

	return packet, nil
}

func (d *DownstreamLocaltcp) DownstreamList() []DownstreamInfo {
	ret := make([]DownstreamInfo, 0)

	d.downstreamsMutex.Lock()
	for _, downstreamInfoTcp := range d.downstreams {
		d := DownstreamInfo{
			downstreamInfoTcp.Name,
			downstreamInfoTcp.Info,
		}
		ret = append(ret, d)
	}
	d.downstreamsMutex.Unlock()

	return ret
}

// startServer is a thread which handles incoming downstream clients and notify parent via channel, lifetime: app
func (d *DownstreamLocaltcp) startServer() (net.Listener, error) {
	log.Info("Start Downstream: LocalTcp on " + d.listenAddr)
	ln, err := net.Listen("tcp", d.listenAddr)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
		return nil, err
	}

	return ln, nil
}

func (d *DownstreamLocaltcp) loop(ln net.Listener, downstreamLocaltcpChannel chan struct{}) {
	n := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Error2: " + err.Error())
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

		d.downstreamsMutex.Lock()
		d.downstreams[name] = downstreamInfoTcp
		d.downstreamsMutex.Unlock()

		// Notify about new downstream
		downstreamLocaltcpChannel <- struct{}{}

		n += 1
	}
}
