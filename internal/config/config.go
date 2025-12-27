package config

import (
	"embed"
	"fmt"
	"os"

	"github.com/jackchuka/gh-brag/internal/data"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "default.yaml"

//go:embed default.yaml
var defaultConfigFS embed.FS

// Theme defines a category and its associated keywords.
type Theme struct {
	Name     string   `yaml:"name"`
	Keywords []string `yaml:"keywords"`
}

// Metrics defines the metrics configuration.
type Metrics struct {
	OwnershipThreshold int                          `yaml:"ownership_threshold"`
	ActionWeights      map[data.EventAction]float64 `yaml:"action_weights"`
	ThemeWeights       map[string]float64           `yaml:"theme_weights"`
}

// Config represents the global configuration for gh-brag.
type Config struct {
	Themes  []Theme `yaml:"themes"`
	Metrics Metrics `yaml:"metrics"`
}

// LoadConfig loads the configuration. It starts with embedded defaults
// and overlays them with the provided YAML file if it exists.
func LoadConfig(path string) (*Config, error) {
	// 1. Start with embedded defaults
	data, err := defaultConfigFS.ReadFile(defaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedded config: %w", err)
	}

	// 2. Overlay with user config if provided
	if path == "" {
		return &cfg, nil
	}

	userData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file %s does not exist", path)
		}
		return nil, err
	}

	if err := yaml.Unmarshal(userData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
	}

	return &cfg, nil
}
