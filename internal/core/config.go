package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration settings, including default model and API keys.
type Config struct {
	DefaultProvider string            `yaml:"default_provider"`
	DefaultModel    string            `yaml:"default_model"`
	APIKeys         map[string]string `yaml:"api_keys"`
}

const configFileName = "config.yaml"

// configPath returns the absolute path to config.yaml in the current working directory.
func configPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to determine working directory: %w", err)
	}
	return filepath.Join(cwd, configFileName), nil
}

// defaultConfigComment is prepended to new config files so users can edit keys directly.
const defaultConfigComment = `# aifiler configuration
# Edit API keys here directly, or run: aifiler set "<provider>"
# Supported providers: openai, anthropic, gemini, ollama, vercel
#
`

// LoadOrDefault attempts to load the configuration from config.yaml in the cwd.
// If the file doesn't exist, it returns a default configuration without error.
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

// InitDefault creates a default config.yaml in the cwd if one does not already exist.
func InitDefault() (string, error) {
	path, err := configPath()
	if err != nil {
		return "", fmt.Errorf("failed to determine config path: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return path, writeConfig(Default(), path)
}

// Save persists the provided configuration back to config.yaml in the cwd.
func Save(cfg Config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", fmt.Errorf("failed to determine config path: %w", err)
	}
	return path, writeConfig(cfg, path)
}

// ConfigPath returns the resolved path to the config file (for display purposes).
func ConfigPath() string {
	p, _ := configPath()
	return p
}

func writeConfig(cfg Config, path string) error {
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}
	content := defaultConfigComment + strings.TrimSpace(string(data)) + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", path, err)
	}
	return nil
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
