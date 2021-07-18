package model

type CmdArgument map[string]string
type CmdResponse map[string]string

type CommandBase struct {
	ComputerId string      `json:"computerid"`
	PacketId   string      `json:"packetid"`
	Command    string      `json:"command"`
	Arguments  CmdArgument `json:"arguments"`
	Response   CmdResponse `json:"response"`
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
