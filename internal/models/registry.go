package models

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultRegistryPath = "assets/models/registry.yaml"

type Registry struct {
	Providers map[string]Provider `yaml:"providers"`
}

type Provider struct {
	Default string   `yaml:"default"`
	Models  []string `yaml:"models"`
}

func LoadRegistry(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, err
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return Registry{}, err
	}
	if reg.Providers == nil {
		reg.Providers = map[string]Provider{}
	}
	return reg, nil
}

func (r Registry) DefaultModelForProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return ""
	}
	entry, ok := r.Providers[provider]
	if !ok {
		return ""
	}
	return entry.Default
}

func (r Registry) Print(filterProvider string) {
	header := "Curated model registry"
	fmt.Println(header)
	filterProvider = strings.ToLower(strings.TrimSpace(filterProvider))
	for name, provider := range r.Providers {
		if filterProvider != "" && name != filterProvider {
			continue
		}
		fmt.Printf("\n%s\n", strings.ToUpper(name))
		fmt.Printf("  default: %s\n", provider.Default)
		for _, model := range provider.Models {
			fmt.Printf("  - %s\n", model)
		}
	}
}
