package model

import (
	"fmt"
	"strconv"
)

type CmdArgument map[string]string
type CmdResponse map[string]string

type CommandBase struct {
	ComputerId string      `json:"computerid"`
	PacketId   string      `json:"packetid"`
	Command    string      `json:"command"`
	Arguments  CmdArgument `json:"arguments"`
	Response   CmdResponse `json:"response"`
}

func MakeCmdArgumentFrom(cmdArgument CmdArgument) (string, []string, error) {
	args := make([]string, 0)

	executable, ok := cmdArgument["executable"]
	if !ok {
		return "", nil, fmt.Errorf("No executable given")
	}

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := "param" + nr
		_, ok := cmdArgument[key]
		if ok {
			fmt.Println("Append: " + cmdArgument[key])
			args = append(args, cmdArgument[key])
		} else {
			break
		}
		n = n + 1
	}

	return executable, args, nil
}

func NewCommand(command string, computerId string, packetId string, arguments CmdArgument, response CmdResponse) CommandBase {
	c := CommandBase{
		computerId,
		packetId,
		command,
		arguments,
		response,
	}
	return c
}
