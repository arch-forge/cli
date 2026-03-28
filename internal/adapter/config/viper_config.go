// Package config implements the ConfigReader port using YAML file I/O.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"gopkg.in/yaml.v3"
)

// ViperConfig implements port.ConfigReader using isolated YAML file operations.
type ViperConfig struct{}

// NewViperConfig constructs a new ViperConfig.
func NewViperConfig() *ViperConfig {
	return &ViperConfig{}
}

// Read parses archforge.yaml at the given file path.
// Returns domain.ErrProjectNotFound if the file does not exist.
// Returns domain.ErrInvalidConfig if the file cannot be parsed.
func (vc *ViperConfig) Read(path string) (*port.ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read config: %w", domain.ErrProjectNotFound)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg port.ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", domain.ErrInvalidConfig)
	}

	return &cfg, nil
}

// Write serialises cfg and writes it to path, creating the file and any
// intermediate directories if they do not exist.
func (vc *ViperConfig) Write(path string, cfg *port.ProjectConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
