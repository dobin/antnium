package client

import (
	"bytes"
	"net/http"
)

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
