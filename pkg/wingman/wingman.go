package wingman

type Wingman struct {
	wingTcp       *WingTcp
	wingDirectory *WingDirectory
}

func MakeWingman() Wingman {
	wingTcp := MakeWingTcp()
	wingDirectory := MakeWingDirectory()

	w := Wingman{
		&wingTcp,
		&wingDirectory,
	}
	return w
}

func (e *Wingman) StartWingman(proto, data string) {
	if proto == "tcp" {
		e.wingTcp.Start(data)
	}
	if proto == "directory" {
		e.wingDirectory.Start(data)
	}
}

func (e *Wingman) Shutdown() {
	if e.wingTcp != nil {
		e.wingTcp.Shutdown()
	}

}
