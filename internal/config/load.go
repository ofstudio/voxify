package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Load loads configuration from [Default] and environment variables
func Load() (Config, error) {
	c := Default()
	if err := env.Parse(&c); err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return c, nil
}
