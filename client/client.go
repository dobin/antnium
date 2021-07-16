package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dobin/antnium/model"
)

type Client struct {
	port int
}

func NewClient(port int) Client {
	w := Client{port}
	return w
}

func (s Client) Start() {
	fmt.Println("Client")

	s.sendPing()
	command := s.getCommand()
	command.Execute()
	fmt.Println("My response: " + command.Response())
	s.sendCommand(command)
}

func (s Client) sendPing() {
	pingCommand := model.NewCommandPing("0", "ooy!")
	s.sendCommand(pingCommand)
}

func (s Client) getCommand() model.Command {
	resp, err := http.Get("http://localhost:4444/getCommand")
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

	fmt.Println("Response: ")
	fmt.Println(bodyString)

	command := model.JsonToCommand(bodyString)
	return command
}

func (s Client) sendCommand(command model.Command) {
	json := command.Json()
	url := "http://localhost:4444/sendCommand"
	fmt.Println("Sending: " + json)
	var jsonStr = []byte(json)
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
