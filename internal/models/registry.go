package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultRegistryPath = "assets/models/registry.yaml"
const RegistryPathEnvVar = "AIFILER_MODEL_REGISTRY"

// Registry holds the configuration of model providers and their available models.
type Registry struct {
	Providers map[string]Provider `yaml:"providers"`
}

// Provider defines the default model and the list of available models for a provider.
type Provider struct {
	Default string   `yaml:"default"`
	Models  []string `yaml:"models"`
}

// LoadRegistry parses the given YAML registry file path and returns a Registry.
func LoadRegistry(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, fmt.Errorf("failed to read registry file: %w", err)
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return Registry{}, fmt.Errorf("failed to parse registry file: %w", err)
	}
	if reg.Providers == nil {
		reg.Providers = map[string]Provider{}
	}
	return reg, nil
}

// LoadDefaultRegistry resolves the default registry path and loads the Registry.
func LoadDefaultRegistry() (Registry, error) {
	path, err := ResolveRegistryPath(DefaultRegistryPath)
	if err != nil {
		return Registry{}, fmt.Errorf("failed to resolve registry path: %w", err)
	}
	return LoadRegistry(path)
}

// ResolveRegistryPath attempts to locate the specified registry file relative
// to current directory, executable directory, or via an environment variable.
func ResolveRegistryPath(defaultRelativePath string) (string, error) {
	defaultRelativePath = strings.TrimSpace(defaultRelativePath)
	if defaultRelativePath == "" {
		defaultRelativePath = DefaultRegistryPath
	}

	if configured := strings.TrimSpace(os.Getenv(RegistryPathEnvVar)); configured != "" {
		if _, err := os.Stat(configured); err == nil {
			return configured, nil
		}
	}

	candidates := []string{}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, defaultRelativePath))
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, defaultRelativePath),
			filepath.Join(exeDir, "..", defaultRelativePath),
		)
	}

	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("registry file not found (set %s or place %s next to executable)", RegistryPathEnvVar, defaultRelativePath)
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
