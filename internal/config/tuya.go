package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/avagenc/zee-api/internal/tuya"
)

func LoadTuya() (*tuya.Config, error) {
    cfg := &tuya.Config{}
    if err := cleanenv.ReadEnv(cfg); err != nil {
        return nil, fmt.Errorf("failed to load tuya config: %w", err)
    }
    return cfg, nil
}
