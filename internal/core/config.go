package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration settings, including default model and API keys.
type Config struct {
	DefaultProvider string            `yaml:"default_provider"`
	DefaultModel    string            `yaml:"default_model"`
	APIKeys         map[string]string `yaml:"api_keys"`
}

// LoadOrDefault attempts to load the configuration from the config file.
// If it fails or the file doesn't exist, it returns a default configuration.
func LoadOrDefault() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), fmt.Errorf("failed to determine config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("failed to parse config file: %w", err)
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	return cfg, nil
}

// InitDefault creates a default configuration file if one does not already exist.
func InitDefault() (string, error) {
	cfg := Default()
	path, err := configPath()
	if err != nil {
		return "", fmt.Errorf("failed to determine config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize default config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("failed to write config file to %s: %w", path, err)
	}
	return path, nil
}

// Save persists the provided configuration to the designated config file path.
func Save(cfg Config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", fmt.Errorf("failed to determine config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("failed to write config file to %s: %w", path, err)
	}
	return path, nil
}

// Default returns a base Config instance with standard application defaults.
func Default() Config {
	return Config{
		DefaultProvider: "none",
		DefaultModel:    "",
		APIKeys: map[string]string{
			"openai":    "",
			"anthropic": "",
			"gemini":    "",
			"vercel":    "",
			"ollama":    "",
		},
	}
}

func configPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to identify user config directory: %w", err)
	}
	return filepath.Join(baseDir, "aifiler", "config.yaml"), nil
}
