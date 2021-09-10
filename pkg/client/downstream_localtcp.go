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
	downstreamsMutex *sync.Mutex          // protect downstreams (necessary?)
	listener         net.Listener         // TCP server, nil = not connected
}

func MakeDownstreamLocaltcp(listenAddr string) DownstreamLocaltcp {
	// Default
	if listenAddr == "" {
		listenAddr = "localhost:50000"
	}

	u := DownstreamLocaltcp{
		listenAddr:       listenAddr,
		downstreams:      make(DownstreamInfoTcpMap, 0),
		downstreamsMutex: &sync.Mutex{},
		listener:         nil,
	}
	return u
}

// Do handles a incoming packet destined for this downstream (localtcp). StartServer() should have been called first.
func (d *DownstreamLocaltcp) Do(packet model.Packet) (model.Packet, error) {
	d.downstreamsMutex.Lock()
	downstreamInfo, ok := d.downstreams[packet.DownstreamId]
	d.downstreamsMutex.Unlock()
	if !ok {
		return model.Packet{}, fmt.Errorf("Did not find: %s", packet.DownstreamId)
	}

	packet, err := d.doConn(downstreamInfo.conn, packet)
	if err != nil {
		log.Error("Error doing localtcp packet: ", err.Error())
		packet.Response["error"] = err.Error()
	}
	return packet, err
}

// doConn will send a packet to a socket and wait for its response
func (d *DownstreamLocaltcp) doConn(conn net.Conn, packet model.Packet) (model.Packet, error) {
	// Send it to the downstream executor
	packetEncoded, err := downstreamclient.EncodePacket(packet)
	if err != nil {
		return packet, err
	}
	_, err = conn.Write(packetEncoded)
	if err != nil {
		/* // TODO put socket in class? (so we can notify webui)
		d.downstreamsMutex.Lock()
		delete(d.downstreams, packet.DownstreamId)
		d.downstreamsMutex.Unlock()

		// Notify about deleted downstream
		downstreamChangeNotifyChan <- struct{}{}
		*/
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

// StartServer starts the TCP listener
func (d *DownstreamLocaltcp) StartServer(downstreamChangeNotifyChan chan struct{}) error {
	log.Info("Start Downstream: LocalTcp on " + d.listenAddr)
	ln, err := net.Listen("tcp", d.listenAddr)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
		return err
	}
	d.listener = ln

	go d.NewConnectionReceiver(downstreamChangeNotifyChan)
	return nil
}

// NewConnectionReceiver is a Thread which waits for new tcp downstream client connections, adds it to the local db and integrates them via DownstreamManager
func (d *DownstreamLocaltcp) NewConnectionReceiver(downstreamChangeNotifyChan chan struct{}) error {
	if d.listener == nil {
		return fmt.Errorf("Can't loop without connection")
	}

	n := 0
	var err error
	for {
		if d.listener == nil {
			log.Info("Listener nil, shutdown thread")
			break
		}
		var conn net.Conn
		conn, err = d.listener.Accept()
		if err != nil {
			break
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
		downstreamChangeNotifyChan <- struct{}{}

		n += 1
	}

	if err != nil {
		log.Errorf("LocalTcp thread exited, because of: %s", err.Error())
	}

	// Never reached
	return err
}

// DownstreamList returns all actively connected clients
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

// Connected returns true if this tcp server is started
func (d *DownstreamLocaltcp) Started() bool {
	if d.listener == nil {
		return false
	} else {
		return true
	}
}

func (d *DownstreamLocaltcp) ListenAddr() string {
	return d.listenAddr
}

func (d *DownstreamLocaltcp) Shutdown() error {
	d.listener.Close()
	d.listener = nil
	d.downstreams = make(DownstreamInfoTcpMap, 0)
	return nil
}
