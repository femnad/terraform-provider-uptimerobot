package uptimerobot

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type MonitorAlertContact struct {
	ID         string `json:"ID,omitempty"`
	Threshold  int64  `json:"threshold,omitempty"`
	Recurrence int64  `json:"recurrence,omitempty"`
}

type Monitor struct {
	AlertContacts []MonitorAlertContact
	FriendlyName  string `json:"friendly_name,omitempty"`
	ID            int64  `json:"id,omitempty"`
	Interval      int64  `json:"interval,omitempty"`
	Status        int64  `json:"status"`
	Timeout       int64  `json:"timeout,omitempty"`
	Type          int64  `json:"type,omitempty"`
	URL           string `json:"url,omitempty"`
}

type deleteMonitorRequest struct {
	auth
	ID int64 `json:"id"`
}

type getMonitorsResponse struct {
	Monitors []Monitor `json:"monitors"`
}

type getMonitorsRequest struct {
	auth
	Monitors string `json:"monitors"`
}

type createMonitorResponse struct {
	baseResponse
	Monitor struct {
		ID     int64 `json:"id,omitempty"`
		Status int64 `json:"status,omitempty"`
	} `json:"monitor"`
}

type editMonitorResponse struct {
	baseResponse
	Monitor struct {
		ID int64 `json:"id,omitempty"`
	} `json:"monitor"`
}

func (c *Client) getDeleteBody(id int64) (io.Reader, error) {
	req := deleteMonitorRequest{ID: id, auth: auth{ApiKey: c.apiKey}}
	return bufferBody(req)
}

func (c *Client) getMonitorsRequestBody() (io.Reader, error) {
	r := getMonitorsRequest{auth: auth{ApiKey: c.apiKey}}
	return bufferBody(r)
}

func (c *Client) getFilteredMonitorsRequestBody(id int64) (io.Reader, error) {
	filterId := strconv.Itoa(int(id))
	r := getMonitorsRequest{Monitors: filterId, auth: auth{ApiKey: c.apiKey}}
	return bufferBody(r)
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

func SerializeMonitorAlertContacts(contacts []MonitorAlertContact) string {
	var alertContacts []string
	for _, contact := range contacts {
		encoded := fmt.Sprintf("%s_%d_%d", contact.ID, contact.Threshold, contact.Recurrence)
		alertContacts = append(alertContacts, encoded)
	}
	return strings.Join(alertContacts, "-")
}

func (c *Client) newMonitorPayload(monitor Monitor) io.Reader {
	v := c.baseValues()
	v.Add("friendly_name", monitor.FriendlyName)
	v.Add("url", monitor.URL)
	v.Add("type", strconv.FormatInt(monitor.Type, 10))
	if monitor.Interval != 0 {
		v.Add("interval", strconv.FormatInt(monitor.Interval, 10))
	}
	if monitor.Timeout != 0 {
		v.Add("timeout", strconv.FormatInt(monitor.Timeout, 10))
	}
	if len(monitor.AlertContacts) > 0 {
		v.Add("alert_contacts", SerializeMonitorAlertContacts(monitor.AlertContacts))
	}

	return strings.NewReader(v.Encode())
}

func (c *Client) editMonitorPayload(monitor Monitor) io.Reader {
	v := c.baseValues()
	v.Add("id", strconv.FormatInt(monitor.ID, 10))
	v.Add("friendly_name", monitor.FriendlyName)
	v.Add("url", monitor.URL)
	v.Add("status", strconv.FormatInt(monitor.Status, 10))
	if monitor.Interval != 0 {
		v.Add("interval", strconv.FormatInt(monitor.Interval, 10))
	}
	if monitor.Timeout != 0 {
		v.Add("timeout", strconv.FormatInt(monitor.Timeout, 10))
	}
	if len(monitor.AlertContacts) > 0 {
		v.Add("alert_contacts", SerializeMonitorAlertContacts(monitor.AlertContacts))
	}

	return strings.NewReader(v.Encode())
}

func (c *Client) GetMonitors() (out []Monitor, err error) {
	url := fmt.Sprintf("%s/getMonitors", baseURL)
	body, err := c.getMonitorsRequestBody()
	if err != nil {
		return
	}

	var resp getMonitorsResponse
	respBody, err := getRequestResp(url, body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	return resp.Monitors, nil
}

func (c *Client) GetMonitor(id int64) (out Monitor, err error) {
	url := fmt.Sprintf("%s/getMonitors", baseURL)
	body, err := c.getFilteredMonitorsRequestBody(id)
	if err != nil {
		return
	}

	var resp getMonitorsResponse
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

func (c *Client) CreateMonitor(monitor Monitor) (out Monitor, err error) {
	newUrl := fmt.Sprintf("%s/newMonitor", baseURL)
	payload := c.newMonitorPayload(monitor)
	respBody, err := postForm(newUrl, payload)
	if err != nil {
		return
	}

	var resp createMonitorResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	if resp.Stat != okStatus {
		return out, fmt.Errorf("unexpected status `%s` when creating monitor, error type: %s, message: %s",
			resp.Stat, resp.Error.Type, resp.Error.Message)
	}

	out.ID = resp.Monitor.ID
	out.Status = resp.Monitor.Status
	return
}

func (c *Client) UpdateMonitor(monitor Monitor) (out Monitor, err error) {
	existing, err := c.GetMonitor(monitor.ID)
	if err != nil {
		return
	}

	if monitor.Type != existing.Type {
		return out, fmt.Errorf("unable to change monitor type via updating")
	}

	editURL := fmt.Sprintf("%s/editMonitor", baseURL)
	payload := c.editMonitorPayload(monitor)

	respBody, err := postForm(editURL, payload)
	if err != nil {
		return
	}

	var resp editMonitorResponse
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	if resp.Stat != okStatus {
		return out, fmt.Errorf("unexpected status `%s` when creating monitor, error type: %s, message: %s",
			resp.Stat, resp.Error.Type, resp.Error.Message)
	}

	return c.GetMonitor(monitor.ID)
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
