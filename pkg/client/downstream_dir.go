package client

import (
	"fmt"
	"os"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	"github.com/dobin/antnium/pkg/wingman"
	log "github.com/sirupsen/logrus"
)

type DownstreamDirectory struct {
	wingmanData campaign.WingmanData
	directory   string
}

func (dd *DownstreamDirectory) SetDirectory(directory string) {
	// We require a trailing slash for now
	if directory[len(directory)-1:] != "\\" {
		directory += "\\"
	}
	dd.directory = directory
}

func MakeDownstreamDirectory(directory string) DownstreamDirectory {
	// Default
	if directory == "" {
		directory = "C:\\temp\\"
	}

	df := DownstreamDirectory{
		campaign.MakeWingmanData(),
		directory,
	}
	return df
}

func (dd *DownstreamDirectory) Do(packet model.Packet) (model.Packet, error) {
	log.Infof("DownstreamDirectory")

	// Write File
	path := dd.directory + dd.wingmanData.Req()

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
	path = dd.directory + dd.wingmanData.Ans()
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

func (dd *DownstreamDirectory) Start(directory string) error {
	dd.SetDirectory(directory)
	log.Info("Start Downstream: Directory on " + dd.directory)
	return nil
}

func (dd *DownstreamDirectory) Directory() string {
	return dd.directory
}

func (dd *DownstreamDirectory) Started() bool {
	if dd.directory == "" {
		return false
	} else {
		return true
	}
}
