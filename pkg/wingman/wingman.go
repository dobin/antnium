package wingman

import (
	"net"
)

type Wingman struct {
	conn net.Conn
}

func MakeWingman() Wingman {
	executor := Wingman{}
	return executor
}

func (e *Wingman) StartWingman(proto string) {
	if proto == "tcp" {
		wingTcp := MakeWingTcp()
		wingTcp.Start("")
	}
	if proto == "directory" {
		wingDirectory := MakeWingDirectory()
		wingDirectory.Start("")
	}
}
