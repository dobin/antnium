package wingman

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

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

func (e *Wingman) StartWingman(proto, data string) error {
	var err error
	if proto == "tcp" {
		err = e.wingTcp.Start(data)
	} else if proto == "directory" {
		err = e.wingDirectory.Start(data)
	} else {
		return fmt.Errorf("Unknown proto: %s", proto)
	}

	if err != nil {
		log.Errorf("%s", err.Error())
	}
	return nil
}

func (e *Wingman) Shutdown() {
	if e.wingTcp != nil {
		e.wingTcp.Shutdown()
	}

}
