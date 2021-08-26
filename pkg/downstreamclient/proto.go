package downstreamclient

import (
	"encoding/json"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

func EncodePacket(packet model.Packet) ([]byte, error) {
	data, err := json.Marshal(packet)
	if err != nil {
		return data, err
	}

	return data, nil
}

func DecodePacket(jsonStr string) (model.Packet, error) {
	// Decode: JSON to GO
	var packet model.Packet
	err := json.Unmarshal([]byte(jsonStr), &packet)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  jsonStr,
			"error": err,
		}).Info("Error sending packet")

		return packet, err
	}

	return packet, nil
}
