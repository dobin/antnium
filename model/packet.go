package model

type Packet struct {
	ComputerId string `json:"computerid"`
	PacketId   string `json:"packetid"`
	Command    string `json:"command"`
	Response   string `json:"response"`
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
type PacketPing struct {
	Packet
}
