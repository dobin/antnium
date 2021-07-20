package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/dobin/antnium/client"
	log "github.com/sirupsen/logrus"
)

func TestGetRequest(t *testing.T) {
	var err error
	var url string

	// Start server in the background
	s := NewServer("127.0.0.1:55000")
	go s.Serve()

	// Create a default (non authenticated) HTTP client
	unauthHttp := &http.Client{
		Timeout: 1 * time.Second,
	}

	/*
		// Test Admin
		r, _ := http.NewRequest("GET", "http://127.0.0.1:55000/admin/commands", nil)
		resp, err := client.Do(r)
		if err != nil {
			panic(err)
		}
		//assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(body)
		//assert.Equal(t, []byte("abcd"), body)
	*/

	c := client.NewClient()
	c.Campaign.ServerUrl = "http://127.0.0.1:55000"

	// Test Client: No key
	url = c.CommandGetUrl()
	r, _ := http.NewRequest("GET", url, nil)
	resp, err := unauthHttp.Do(r)
	if err != nil {
		t.Errorf("Error accessing server api with url: " + url)
	}
	if resp.StatusCode == 200 {
		t.Errorf("Could access server API though i should not: " + url)
	}

	// Test Client: Correct key
	url = c.CommandGetUrl()
	resp, err = c.HttpGet(url)
	if resp.StatusCode != 200 {
		t.Errorf("Could not access server API for client: " + url)
	}
	log.Println(resp)

	// Test: Static
	/*
		url = c.CommandGetUrl()
		r, _ = http.NewRequest("GET", url, nil)
		resp, err = unauthHttp.Do(r)
		if err != nil {
			t.Errorf("Error accessing static with url: " + url)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Could access static: " + url)
		}*/

}
