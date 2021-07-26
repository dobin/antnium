package model

import (
	"fmt"
	"strconv"
)

type PacketArgument map[string]string
type PacketResponse map[string]string

type Packet struct {
	ComputerId string         `json:"computerid"`
	PacketId   string         `json:"packetid"`
	Command    string         `json:"command"`
	Arguments  PacketArgument `json:"arguments"`
	Response   PacketResponse `json:"response"`
}

func MakePacketArgumentFrom(packetArgument PacketArgument) (string, []string, error) {
	args := make([]string, 0)

	executable, ok := packetArgument["executable"]
	if !ok {
		return "", nil, fmt.Errorf("No executable given")
	}

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := "param" + nr
		_, ok := packetArgument[key]
		if ok {
			fmt.Println("Append: " + packetArgument[key])
			args = append(args, packetArgument[key])
		} else {
			break
		}
		n = n + 1
	}

	return executable, args, nil
}

func NewCommand(command string, computerId string, packetId string, arguments PacketArgument, response PacketResponse) Packet {
	c := Packet{
		computerId,
		packetId,
		command,
		arguments,
		response,
	}
	return c
}
