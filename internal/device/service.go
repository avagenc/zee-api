package device

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/avagenc/zee/internal/domain"
)

type TuyaClient interface {
	QueryDevicesInHome(homeID string) (json.RawMessage, error)
	SendCommands(deviceID string, commands any) (json.RawMessage, error)
	GetMultiChannelName(deviceID string) (json.RawMessage, error)
	GetUserDeviceList(tuyaUID string) ([]domain.Device, error)
}

type service struct {
	tuya TuyaClient
}

func NewService(tuya TuyaClient) *service {
	return &service{tuya: tuya}
}

func (s *service) SendCommands(deviceID string, commands []domain.DataPoint) (json.RawMessage, error) {
	result, err := s.tuya.SendCommands(deviceID, commands)
	if err != nil {
		return nil, fmt.Errorf("failed to send commands: %w", err)
	}
	return result, nil
}

func (s *service) ListByTuyaUID(tuyaUID string) ([]domain.Device, error) {
	devices, err := s.tuya.GetUserDeviceList(tuyaUID)
	if err != nil {
		return nil, err
	}

	if len(devices) == 0 {
		return []domain.Device{}, nil
	}

	if err := s.enrichDevices(devices); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	return devices, nil
}

// Deprecated: Unused. Kept for future reference.
func (s *service) ListByHome(homeID string) ([]domain.Device, error) {
	result, err := s.tuya.QueryDevicesInHome(homeID)
	if err != nil {
		return nil, err
	}

	var devices []domain.Device
	if len(result) > 0 {
		if err := json.Unmarshal(result, &devices); err != nil {
			return nil, fmt.Errorf("failed to decode devices: %w", err)
		}
	}

	if len(devices) == 0 {
		return []domain.Device{}, nil
	}

	if err := s.enrichDevices(devices); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	return devices, nil
}

func (s *service) enrichDevices(devices []domain.Device) error {
	var devicesToEnrich []*domain.Device
	for i := range devices {
		device := &devices[i]
		category := strings.ToLower(device.Category)
		device.CodeNameMapping = []domain.Channel{}

		if (category == "kg" || strings.HasPrefix(category, "cz")) && device.ID != "" {
			devicesToEnrich = append(devicesToEnrich, device)
		}
	}

	if len(devicesToEnrich) > 0 {
		return s.enrichWithChannelNames(devicesToEnrich)
	}
	return nil
}

func (s *service) enrichWithChannelNames(devices []*domain.Device) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(devices))

	for _, device := range devices {
		wg.Add(1)
		go func(device *domain.Device) {
			defer wg.Done()

			result, err := s.tuya.GetMultiChannelName(device.ID)
			if err != nil {
				errs <- fmt.Errorf("failed to get channel name for device %s: %w", device.ID, err)
				return
			}

			var channels []domain.Channel
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
