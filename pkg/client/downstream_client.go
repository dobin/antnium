package client

import (
	"github.com/dobin/antnium/pkg/executor"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type DownstreamClient struct {
	packetExecutor executor.PacketExecutor
}

func MakeDownstreamClient() DownstreamClient {
	u := DownstreamClient{
		executor.MakePacketExecutor(),
	}
	return u
}

func (d *DownstreamClient) do(packet model.Packet) (model.Packet, error) {
	err := d.packetExecutor.Execute(&packet)
	if err != nil {
		log.WithFields(log.Fields{
			"packet": packet,
			"error":  err,
		}).Info("Error executing packet")
		packet.Response["error"] = err.Error()
		return packet, err
	}
	return packet, nil
}
