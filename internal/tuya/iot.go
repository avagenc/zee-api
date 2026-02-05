package tuya

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	devicesEndpoint    = "/v1.0/devices"
	cloudThingEndpoint = "/v2.0/cloud/thing"
	homeEndpoint       = "/v1.0/homes"
)

func (c *Client) QueryProperties(deviceID string) (json.RawMessage, error) {
	path := fmt.Sprintf("%s/%s/shadow/properties", cloudThingEndpoint, deviceID)
	tuyaReq := Request{
		Method:  http.MethodGet,
		URLPath: path,
	}
	resp, err := c.doIoTRequest(tuyaReq)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (c *Client) SendCommands(deviceID string, commands any) (json.RawMessage, error) {
	path := fmt.Sprintf("%s/%s/commands", devicesEndpoint, deviceID)
	bodyBytes, err := json.Marshal(struct {
		Commands any `json:"commands"`
	}{
		Commands: commands,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command payload: %w", err)
	}

	tuyaReq := Request{
		Method:  http.MethodPost,
		URLPath: path,
		Body:    string(bodyBytes),
	}
	resp, err := c.doIoTRequest(tuyaReq)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (c *Client) GetMultiChannelName(deviceID string) (json.RawMessage, error) {
	path := fmt.Sprintf("%s/%s/multiple-names", devicesEndpoint, deviceID)
	tuyaReq := Request{
		Method:  http.MethodGet,
		URLPath: path,
	}
	resp, err := c.doIoTRequest(tuyaReq)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (c *Client) QueryDevicesInHome(homeID string) (json.RawMessage, error) {
	path := fmt.Sprintf("%s/%s/devices", homeEndpoint, homeID)
	tuyaReq := Request{
		Method:  http.MethodGet,
		URLPath: path,
	}
	resp, err := c.doIoTRequest(tuyaReq)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}
