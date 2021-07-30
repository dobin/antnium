package client

import (
	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

type Downstream struct {
	channel chan model.Packet

	packetExecutor executor.PacketExecutor
}

func MakeDownstream() Downstream {
	u := Downstream{
		make(chan model.Packet),
		executor.MakePacketExecutor(),
	}
	return u
}

func (d Downstream) start() {
	for {
		packet := <-d.channel

		err := d.packetExecutor.Execute(&packet)
		if err != nil {
			log.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Info("Error executing packet")
		}

		// Always send response, as it is syncronous
		d.channel <- packet
	}
}
