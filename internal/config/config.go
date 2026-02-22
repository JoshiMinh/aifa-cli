package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProvider string            `yaml:"default_provider"`
	DefaultModel    string            `yaml:"default_model"`
	APIKeys         map[string]string `yaml:"api_keys"`
}

func LoadOrDefault() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	return cfg, nil
}

func InitDefault() (string, error) {
	cfg := Default()
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func Save(cfg Config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func Default() Config {
	return Config{
		DefaultProvider: "none",
		DefaultModel:    "",
		APIKeys: map[string]string{
			"openai":    "",
			"anthropic": "",
			"google":    "",
		},
	}
}

func configPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "aifiler", "config.yaml"), nil
}
