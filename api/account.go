package uptimerobot

import (
	"encoding/json"
	"fmt"
)

type Account struct {
	Email           string `json:"email"`
	MonitorLimit    int64  `json:"monitor_limit"`
	MonitorInterval int64  `json:"monitor_interval"`
	UpMonitors      int64  `json:"up_monitors"`
	DownMonitors    int64  `json:"down_monitors"`
	PausedMonitors  int64  `json:"paused_monitors"`
}

type accountResp struct {
	baseResponse
	Account Account `json:"account"`
}

func (c *Client) GetAccount() (acc Account, err error) {
	url := fmt.Sprintf("%s/getAccountDetails", baseURL)
	body, err := c.getAuthBody()
	if err != nil {
		return
	}

	respBody, err := getRequestResp(url, body)
	if err != nil {
		return
	}

	var resp accountResp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	if resp.Stat != okStatus {
		return acc, fmt.Errorf("unexpected status %s when getting account details", resp.Stat)
	}

	return resp.Account, nil
}
