package model

import (
	"fmt"
	"strconv"
)

type PacketArgument map[string]string
type PacketResponse map[string]string

type Packet struct {
	ComputerId   string         `json:"computerid"`
	PacketId     string         `json:"packetid"`
	PacketType   string         `json:"packetType"`
	Arguments    PacketArgument `json:"arguments"`
	Response     PacketResponse `json:"response"`
	DownstreamId string         `json:"downstreamId"`
}

func NewPacket(packetType string, computerId string, packetId string, arguments PacketArgument, response PacketResponse) Packet {
	c := Packet{
		ComputerId:   computerId,
		PacketId:     packetId,
		PacketType:   packetType,
		Arguments:    arguments,
		Response:     response,
		DownstreamId: "client", // sane default for now
	}
	return c
}

func AddArrayToResponse(key string, data []string, response PacketResponse) {
	for i, value := range data {
		dictKey := key + strconv.Itoa(i)
		response[dictKey] = value
	}
}

func ResponseToArray(baseKey string, response PacketResponse) []string {
	data := make([]string, 0)

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := baseKey + nr
		_, ok := response[key]
		if ok {
			data = append(data, response[key])
		} else {
			break
		}
		n = n + 1
	}

	return data
}

func MakePacketArgumentFrom(packetArgument PacketArgument) (string, []string, error) {
	args := make([]string, 0)

	executable, ok := packetArgument["executable"]
	if !ok {
		return "", nil, fmt.Errorf("missing argument 'executable'")
	}

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := "param" + nr
		_, ok := packetArgument[key]
		if ok {
			args = append(args, packetArgument[key])
		} else {
			break
		}
		n = n + 1
	}

	return executable, args, nil
}
