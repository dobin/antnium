package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	resp, err := http.Get("http://localhost:4444/test")
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
}
