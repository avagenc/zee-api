package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type DataPoint struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}

type Channel struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type Device struct {
	ID              string      `json:"id"`
	Category        string      `json:"category"`
	Name            string      `json:"name"`
	Status          []DataPoint `json:"status"`
	CodeNameMapping []Channel   `json:"code_name_mapping"`
}

type TuyaClient interface {
	QueryDevicesInHome(homeID string) (json.RawMessage, error)
	SendCommands(deviceID string, commands any) (json.RawMessage, error)
	GetMultiChannelName(deviceID string) (json.RawMessage, error)
}

type service struct {
	tuya TuyaClient
}

func NewService(tuya TuyaClient) *service {
	return &service{tuya: tuya}
}

func (s *service) SendCommands(deviceID string, commands []DataPoint) (json.RawMessage, error) {
	result, err := s.tuya.SendCommands(deviceID, commands)
	if err != nil {
		return nil, fmt.Errorf("failed to send commands: %w", err)
	}
	return result, nil
}

func (s *service) ListByHome(homeID string) ([]Device, error) {
	result, err := s.tuya.QueryDevicesInHome(homeID)
	if err != nil {
		return nil, err
	}

	var devices []Device
	if len(result) > 0 {
		if err := json.Unmarshal(result, &devices); err != nil {
			return nil, fmt.Errorf("failed to decode devices: %w", err)
		}
	}

	if len(devices) == 0 {
		return []Device{}, nil
	}

	var devicesToEnrich []*Device
	for i := range devices {
		device := &devices[i]
		category := strings.ToLower(device.Category)
		device.CodeNameMapping = []Channel{}

		if (category == "kg" || strings.HasPrefix(category, "cz")) && device.ID != "" {
			devicesToEnrich = append(devicesToEnrich, device)
		}
	}

	if len(devicesToEnrich) > 0 {
		if err := s.enrichWithChannelNames(devicesToEnrich); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return devices, nil
}

func (s *service) enrichWithChannelNames(devices []*Device) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(devices))

	for _, device := range devices {
		wg.Add(1)
		go func(device *Device) {
			defer wg.Done()

			result, err := s.tuya.GetMultiChannelName(device.ID)
			if err != nil {
				errs <- fmt.Errorf("failed to get channel name for device %s: %w", device.ID, err)
				return
			}

			var channels []Channel
			if len(result) > 0 {
				if err := json.Unmarshal(result, &channels); err != nil {
					errs <- fmt.Errorf("failed to decode channels for device %s: %w", device.ID, err)
					return
				}
			}
			device.CodeNameMapping = channels
		}(device)
	}

	wg.Wait()
	close(errs)

	var allErrors []string
	for err := range errs {
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("encountered %d error(s): %s", len(allErrors), strings.Join(allErrors, "; "))
	}

	return nil
}
