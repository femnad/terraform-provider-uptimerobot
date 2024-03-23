package uptimerobot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

const (
	alertContactInitialStatus = 0
)

type AlertContact struct {
	ID           string `json:"id,omitempty"`
	FriendlyName string `json:"friendly_name,omitempty"`
	Type         int64  `json:"type,omitempty"`
	Status       int64  `json:"status,omitempty"`
	Value        string `json:"value,omitempty"`
}

type alertContactsResponse struct {
	baseResponse
	AlertContacts []AlertContact `json:"alert_contacts,omitempty"`
}

type alertContactResponse struct {
	baseResponse
	AlertContact struct {
		ID int64 `json:"id"`
	} `json:"alertcontact"`
}

func alertContactMapLookup(attr string, designator int64, mapping map[int64]string) (string, error) {
	str, ok := mapping[designator]
	if !ok {
		return "", fmt.Errorf("no alert contact %s exists for type designator %d", attr, designator)
	}

	return str, nil
}

func alertContactReverseMapLookup(attr, strVal string, mapping map[int64]string) (int64, error) {
	for k, v := range mapping {
		if v == strVal {
			return k, nil
		}
	}

	return 0, fmt.Errorf("no alert contact %s designator exists for string %s", attr, strVal)
}

func AlertContactStatusToString(status int64) (string, error) {
	return alertContactMapLookup("status", status, AlertContactStatuses)
}

func AlertContactTypeToString(contactType int64) (string, error) {
	return alertContactMapLookup("type", contactType, AlertContactTypes)
}

func AlertContactStatusToDesignator(contactType string) (int64, error) {
	return alertContactReverseMapLookup("status", contactType, AlertContactStatuses)
}

func AlertContactTypeToDesignator(contactType string) (int64, error) {
	return alertContactReverseMapLookup("type", contactType, AlertContactTypes)
}

func (c *Client) baseValues() url.Values {
	values := url.Values{}
	values.Set("api_key", c.apiKey)
	values.Set("format", "json")
	return values
}

func (c *Client) newAlertContactPayload(contact AlertContact) io.Reader {
	v := c.baseValues()
	v.Set("friendly_name", contact.FriendlyName)
	v.Set("type", strconv.FormatInt(contact.Type, 10))
	v.Set("value", contact.Value)
	return strings.NewReader(v.Encode())
}

func (c *Client) editContactPayload(contact AlertContact) io.Reader {
	v := c.baseValues()
	v.Set("id", contact.ID)
	v.Set("friendly_name", contact.FriendlyName)
	v.Set("value", contact.Value)
	return strings.NewReader(v.Encode())
}

func (c *Client) deleteAlertContactPayload(contact AlertContact) io.Reader {
	v := c.baseValues()
	v.Set("id", contact.ID)
	return strings.NewReader(v.Encode())
}

func (c *Client) getAlertContactPayload(contact AlertContact) io.Reader {
	v := c.baseValues()
	v.Set("alert_contacts", contact.ID)
	return strings.NewReader(v.Encode())
}

func listAlertContacts(methodURL string, payload io.Reader) (resp alertContactsResponse, err error) {
	respBody, err := postForm(methodURL, payload)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	if resp.Stat != okStatus {
		return resp, fmt.Errorf("unexpected status `%s` when creating alert contact, error type: %s, message: %s",
			resp.Stat, resp.Error.Type, resp.Error.Message)
	}

	return resp, nil
}

func processAlertContact(methodURL string, payload io.Reader) (resp alertContactResponse, err error) {
	respBody, err := postForm(methodURL, payload)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}

	if resp.Stat != okStatus {
		return resp, fmt.Errorf("unexpected status `%s` when creating alert contact, error type: %s, message: %s",
			resp.Stat, resp.Error.Type, resp.Error.Message)
	}

	return resp, nil
}

func (c *Client) GetAlertContacts() (contacts []AlertContact, err error) {
	getURL := fmt.Sprintf("%s/getAlertContacts", baseURL)
	payload := strings.NewReader(c.baseValues().Encode())
	resp, err := listAlertContacts(getURL, payload)
	if err != nil {
		return
	}

	return resp.AlertContacts, err
}

func (c *Client) GetAlertContact(contact AlertContact) (AlertContact, error) {
	getURL := fmt.Sprintf("%s/getAlertContacts", baseURL)
	payload := c.getAlertContactPayload(contact)
	resp, err := listAlertContacts(getURL, payload)
	if err != nil {
		return contact, err
	}

	if len(resp.AlertContacts) != 1 {
		return contact, fmt.Errorf("unexpected number of alert contacts return for contact ID: %s", contact.ID)
	}

	return resp.AlertContacts[0], nil
}

func (c *Client) CreateAlertContact(contact AlertContact) (out AlertContact, err error) {
	newURL := fmt.Sprintf("%s/newAlertContact", baseURL)
	payload := c.newAlertContactPayload(contact)
	resp, err := processAlertContact(newURL, payload)
	if err != nil {
		return
	}

	out.ID = strconv.FormatInt(resp.AlertContact.ID, 10)
	out.Status = alertContactInitialStatus
	return
}

func (c *Client) UpdateAlertContact(contact AlertContact) (out AlertContact, err error) {
	fetchedContact, err := c.GetAlertContact(contact)
	if err != nil {
		return
	}

	editURL := fmt.Sprintf("%s/editAlertContact", baseURL)
	payload := c.editContactPayload(contact)
	resp, err := processAlertContact(editURL, payload)
	if err != nil {
		return
	}

	out.ID = strconv.FormatInt(resp.AlertContact.ID, 10)
	out.FriendlyName = contact.FriendlyName
	out.Status = fetchedContact.Status
	out.Value = contact.Value
	return
}

func (c *Client) DeleteAlertContact(contact AlertContact) (err error) {
	deleteURL := fmt.Sprintf("%s/deleteAlertContact", baseURL)
	payload := c.deleteAlertContactPayload(contact)
	_, err = processAlertContact(deleteURL, payload)
	return
}
