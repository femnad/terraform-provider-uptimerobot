package uptimerobot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	baseURL         = "https://api.uptimerobot.com/v2"
	jsonContentType = "application/json"
)

var (
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

type Monitor struct {
	ID           int64  `json:"id,omitempty"`
	FriendlyName string `json:"friendly_name,omitempty"`
	URL          string `json:"url,omitempty"`
	Type         int64  `json:"type,omitempty"`
	Interval     int64  `json:"interval,omitempty"`
	Timeout      int64  `json:"timeout,omitempty"`
}

type monitorRequest struct {
	Monitor
	auth
}

type deleteRequest struct {
	ID int64 `json:"id"`
	auth
}

type monitorsResponse struct {
	Monitors []Monitor `json:"monitors"`
}

type monitorsRequest struct {
	Monitors string `json:"monitors"`
	auth
}

type createResponse struct {
	ID     int64 `json:"id,omitempty"`
	Status int64 `json:"status,omitempty"`
}

type monitorCreateResponse struct {
	Monitor createResponse `json:"monitor"`
}

func MonitorTypeToInt(strType string) (int64, error) {
	intType, ok := MonitorTypes[strType]
	if !ok {
		return 0, fmt.Errorf("unable to corresponding monitor type for %s", strType)
	}

	return intType, nil
}

func MonitorTypeToStr(intType int64) (string, error) {
	for s, i := range MonitorTypes {
		if i == intType {
			return s, nil
		}
	}

	return "", fmt.Errorf("unable to corresponding monitor type for %d", intType)
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

func (c *Client) getMonitorBody(monitor Monitor) (io.Reader, error) {
	r := monitorRequest{Monitor: monitor, auth: auth{ApiKey: c.apiKey}}
	return bufferBody(r)
}

func (c *Client) getDeleteBody(id int64) (io.Reader, error) {
	req := deleteRequest{ID: id, auth: auth{ApiKey: c.apiKey}}
	return bufferBody(req)
}

func (c *Client) getMonitorRequestBody(id int64) (io.Reader, error) {
	filterId := strconv.Itoa(int(id))
	r := monitorsRequest{Monitors: filterId, auth: auth{ApiKey: c.apiKey}}
	return bufferBody(r)
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

func (c *Client) createOrUpdate(monitor Monitor, url string) (out Monitor, err error) {
	body, err := c.getMonitorBody(monitor)
	if err != nil {
		return
	}

	var resp monitorCreateResponse
	respBody, err := getRequestResp(url, body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	out = monitor
	out.ID = resp.Monitor.ID
	return
}

func (c *Client) CreateMonitor(monitor Monitor) (out Monitor, err error) {
	url := fmt.Sprintf("%s/newMonitor", baseURL)
	out, err = c.createOrUpdate(monitor, url)
	if err != nil {
		return
	}

	return c.createOrUpdate(monitor, url)
}

func (c *Client) GetMonitor(id int64) (out Monitor, err error) {
	url := fmt.Sprintf("%s/getMonitors", baseURL)
	body, err := c.getMonitorRequestBody(id)
	if err != nil {
		return
	}

	var resp monitorsResponse
	respBody, err := getRequestResp(url, body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	for _, monitor := range resp.Monitors {
		if id == monitor.ID {
			return monitor, nil
		}
	}

	return out, fmt.Errorf("unable to find monitor with id %d", id)
}

func (c *Client) UpdateMonitor(monitor Monitor) (out Monitor, err error) {
	existing, err := c.GetMonitor(monitor.ID)
	if err != nil {
		return
	}

	if monitor.Type != existing.Type {
		return out, fmt.Errorf("unable to change monitor type via updating")
	}

	url := fmt.Sprintf("%s/editMonitor", baseURL)
	return c.createOrUpdate(monitor, url)
}

func (c *Client) DeleteMonitor(id int64) (err error) {
	url := fmt.Sprintf("%s/deleteMonitor", baseURL)
	body, err := c.getDeleteBody(id)
	if err != nil {
		return
	}

	_, err = getRequestResp(url, body)
	return
}

func New(apiKey string) (*Client, error) {
	return &Client{apiKey: apiKey}, nil
}
