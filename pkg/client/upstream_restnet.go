package client

import (
	"bytes"
	"net/http"
)

func (u *UpstreamRest) PacketGetUrl() string {
	return u.campaign.ServerUrl + u.campaign.PacketGetPath + u.config.ClientId
}

func (u *UpstreamRest) PacketSendUrl() string {
	return u.campaign.ServerUrl + u.campaign.PacketSendPath
}

func (u *UpstreamRest) HttpGet(url string) (*http.Response, error) {
	client := u.httpClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(u.campaign.AuthHeader, u.campaign.ApiKey)
	req.Header.Set("User-Agent", u.campaign.UserAgent)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *UpstreamRest) HttpPost(url string, data *bytes.Reader) (*http.Response, error) {
	client := u.httpClient
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set(u.campaign.AuthHeader, u.campaign.ApiKey)
	req.Header.Set("User-Agent", u.campaign.UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
