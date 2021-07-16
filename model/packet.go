package model

type Packet struct {
	Command  string `json:"command"`
	Response string `json:"response"`
}

type PacketExec struct {
	Packet
	Arguments []string `json:"arguments"`
}

type PacketTest struct {
	Packet
	Arguments []string `json:"arguments"`
}

type PacketInfo struct {
	Packet
}
