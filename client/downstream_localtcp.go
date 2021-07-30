package client

import (
	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamLocaltcp struct {
	channel chan model.Packet

	packetExecutor executor.PacketExecutor
}

func MakeDownstreamLocaltcp() DownstreamLocaltcp {
	u := DownstreamLocaltcp{
		make(chan model.Packet),
		executor.MakePacketExecutor(),
	}
	return u
}

func (d DownstreamLocaltcp) start() {
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
