package uptimerobot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL         = "https://api.uptimerobot.com/v2"
	jsonContentType = "application/json"
)

type AlertContact struct {
	ID           string `json:"id,omitempty"`
	FriendlyName string `json:"friendly_name,omitempty"`
	Type         int64  `json:"type,omitempty"`
	Status       int64  `json:"status,omitempty"`
	Value        string `json:"value,omitempty"`
}

type alertContactsResponse struct {
	AlertContacts []AlertContact `json:"alert_contacts,omitempty"`
}

type auth struct {
	ApiKey string `json:"api_key"`
}

type Client struct {
	apiKey string
}

func (c Client) getBody() (io.Reader, error) {
	a := auth{ApiKey: c.apiKey}

	out, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(out), nil
}

func (c Client) GetAlertContacts() ([]AlertContact, error) {
	url := fmt.Sprintf("%s/getAlertContacts", baseURL)
	body, err := c.getBody()
	if err != nil {
		return nil, err
	}

	var respBody []byte
	resp, err := http.Post(url, jsonContentType, body)
	readBody := func() ([]byte, error) {
		if resp == nil {
			return nil, nil
		}
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		return respBody, err
	}

	if err != nil {
		return nil, fmt.Errorf("error getting alert contacts: body %s, error %v", respBody, err)
	}

	if resp.StatusCode >= 400 {
		respBody, err = readBody()
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error getting alert contacts: status code: %d, body %s, error %v",
			resp.StatusCode, respBody, err)
	}

	respBody, err = readBody()
	if err != nil {
		return nil, err
	}

	var acResp alertContactsResponse
	err = json.Unmarshal(respBody, &acResp)
	return acResp.AlertContacts, err
}

func New(apiKey string) (*Client, error) {
	return &Client{apiKey: apiKey}, nil
}
