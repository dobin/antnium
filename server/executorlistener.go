package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

func StartServer() {
	ln, err := net.Listen("tcp", "127.0.0.1:50000")
	if err != nil {
		log.Error("Error: " + err.Error())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Error: " + err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Make a example packet the client should receive
	arguments := make(model.PacketArgument)
	arguments["executable"] = "cmd"
	arguments["param0"] = "/c"
	arguments["param1"] = "hostname"
	response := make(model.PacketResponse)
	packet := model.NewPacket("exec", "c-1", "p-1", arguments, response)
	fmt.Printf("Send: %v\n", packet)

	packetEncoded, err := executor.EncodePacket(packet)
	if err != nil {
		log.Error("Error: ", err.Error())
	}
	conn.Write(packetEncoded)
	conn.Write([]byte("\n"))

	fmt.Println("Receive:")
	jsonStr, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Error("Could not read: " + err.Error())
	}

	fmt.Println("Jsonstr: " + jsonStr)
	packetDecoded, err := executor.DecodePacket(jsonStr)
	if err != nil {
		log.Error("Error: ", err.Error())
	}
	fmt.Printf("%v\n", packetDecoded)

}
