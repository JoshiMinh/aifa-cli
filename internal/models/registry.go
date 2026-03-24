package models

import (
	"strings"
)

// Registry holds the configuration of model providers and their available models.
type Registry struct {
	Providers map[string]Provider `yaml:"providers"`
}

// Provider defines the default model and the list of available models for a provider.
type Provider struct {
	Default string   `yaml:"default"`
	Models  []string `yaml:"models"`
}

func (r Registry) DefaultModelForProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || r.Providers == nil {
		return ""
	}
	entry, ok := r.Providers[provider]
	if !ok {
		return ""
	}
	return entry.Default
}
