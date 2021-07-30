package client

import (
	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamClient struct {
	channel chan model.Packet

	packetExecutor executor.PacketExecutor
}

func MakeDownstreamClient() DownstreamClient {
	u := DownstreamClient{
		make(chan model.Packet),
		executor.MakePacketExecutor(),
	}
	return u
}

func (d *DownstreamClient) start() {
	log.Info("Start Downstream: Client")

	for {
		packet := <-d.channel

		err := d.packetExecutor.Execute(&packet)
		if err != nil {
			log.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Info("Error executing packet")
			packet.Response["error"] = err.Error()
		}

		// Always send response, as it is syncronous
		d.channel <- packet
	}
}
