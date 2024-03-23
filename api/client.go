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
	okStatus        = "ok"
)

var (
	AlertContactStatuses = map[int64]string{
		0: "not activated",
		1: "paused",
		2: "active",
	}
	AlertContactTypes = map[int64]string{
		1:  "sms",
		2:  "e-mail",
		3:  "twitter",
		4:  "web-hook",
		5:  "pushbullet",
		6:  "zapier",
		7:  "pro-sms",
		8:  "pushover",
		9:  "slack",
		10: "voice-call",
		11: "splunk",
		12: "pagerduty",
		13: "opsgenie",
		14: "ms-teams",
		15: "google-chat",
		16: "discord",
	}
	MonitorTypes = map[string]int64{
		"http":      1,
		"keyword":   2,
		"ping":      3,
		"port":      4,
		"heartbeat": 5,
	}
)

type auth struct {
	ApiKey string `json:"api_key"`
}

type baseResponse struct {
	Stat string `json:"stat"`
}

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

type Client struct {
	apiKey string
}

func bufferBody(a any) (io.Reader, error) {
	out, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(out), nil
}

func (c *Client) getAuthBody() (io.Reader, error) {
	a := auth{ApiKey: c.apiKey}
	return bufferBody(a)
}

func readRespBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return respBody, err
}

func getRequestResp(url string, body io.Reader) ([]byte, error) {
	resp, err := http.Post(url, jsonContentType, body)
	if err != nil {
		return nil, fmt.Errorf("error getting alert contacts: error %v", err)
	}

	var respBody []byte
	if resp.StatusCode >= 400 {
		respBody, err = readRespBody(resp)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("unexpected response: status code %d, body %s, error %v",
			resp.StatusCode, respBody, err)
	}

	return readRespBody(resp)
}

func (c *Client) GetAlertContacts() ([]AlertContact, error) {
	url := fmt.Sprintf("%s/getAlertContacts", baseURL)
	body, err := c.getAuthBody()
	if err != nil {
		return nil, err
	}

	var resp alertContactsResponse
	respBody, err := getRequestResp(url, body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBody, &resp)
	return resp.AlertContacts, err
}

func New(apiKey string) (*Client, error) {
	return &Client{apiKey: apiKey}, nil
}
