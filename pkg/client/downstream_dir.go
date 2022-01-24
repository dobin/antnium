package client

import (
	"fmt"
	"os"
	"time"

	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/wingman"
	log "github.com/sirupsen/logrus"
)

type DownstreamDirectory struct {
	directory string
}

func MakeDownstreamDirectory(directory string) DownstreamDirectory {
	df := DownstreamDirectory{
		"ipc/",
	}
	return df
}

func (df *DownstreamDirectory) Do(packet model.Packet) (model.Packet, error) {
	log.Infof("DownstreamDirectory")

	// Write File
	path := df.directory + "1.dwn"

	// Send it to the downstream executor
	packetEncoded, err := wingman.EncodePacket(packet)
	if err != nil {
		return packet, err
	}

	err = os.WriteFile(path, packetEncoded, 0644)
	if err != nil {
		return packet, err
	}

	// Read answer
	path = df.directory + "1.pu"
	max := 50
	for {
		if max <= 0 {
			break
		}

		data, err := os.ReadFile(path)
		if err == nil {
			// Always delete file
			os.Remove(path)

			packetAnswer, err := wingman.DecodePacket(string(data))
			if err != nil {
				return packetAnswer, err
			}

			// Return original packet!
			return packetAnswer, nil
		}

		time.Sleep(20 * time.Millisecond)
		max -= 1
	}

	// Not reached
	return packet, fmt.Errorf("Answer not received in time")
}
