package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"aifiler/internal/config"
	"aifiler/internal/llm"
	"aifiler/internal/models"
)

func (a *App) runList(ctx context.Context) int {
	registry, err := models.LoadDefaultRegistry()
	if err != nil {
		errorStyle.Printf("failed to load model registry: %v\n", err)
		return 1
	}
	cfg, _ := config.LoadOrDefault()

	headerStyle.Println("Available providers and models")
	registry.Print("")

	fmt.Println()
	headerStyle.Println("Configured API keys")
	providers := make([]string, 0, len(registry.Providers)+len(cfg.APIKeys))
	providerSet := map[string]struct{}{}
	for provider := range registry.Providers {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" || provider == "none" {
			continue
		}
		providerSet[provider] = struct{}{}
	}
	for provider := range cfg.APIKeys {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" {
			continue
		}
		providerSet[provider] = struct{}{}
	}
	for provider := range providerSet {
		providers = append(providers, provider)
	}
	sort.Strings(providers)

	if len(providers) == 0 {
		fmt.Println("- none")
	} else {
		for _, provider := range providers {
			status := "not-set"
			if strings.TrimSpace(cfg.APIKeys[provider]) != "" {
				status = "set"
			}
			fmt.Printf("- %s: %s\n", provider, status)
		}
	}

	fmt.Println()
	fmt.Printf("default_provider: %s\n", cfg.DefaultProvider)
	fmt.Printf("default_model: %s\n", cfg.DefaultModel)

	ollamaModels, err := llm.DetectOllamaModels(ctx)
	if err == nil && len(ollamaModels) > 0 {
		fmt.Println()
		headerStyle.Println("Detected local Ollama models")
		for _, model := range ollamaModels {
			fmt.Printf("  - %s\n", model)
		}
	}

	vercelModels, err := llm.DetectVercelModels(ctx, cfg.APIKeys["vercel"], "")
	if err == nil && len(vercelModels) > 0 {
		fmt.Println()
		headerStyle.Println("Detected Vercel AI Gateway models")
		for _, model := range vercelModels {
			fmt.Printf("  - %s\n", model)
		}
	}
	return 0
}

func (a *App) runSet(args []string) int {
	if len(args) < 2 {
		errorStyle.Println("Usage: aifiler set \"provider\" \"api key\"")
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	apiKey := strings.TrimSpace(args[1])
	if provider == "" || apiKey == "" {
		errorStyle.Println("provider and api key cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	cfg.APIKeys[provider] = apiKey
	if strings.TrimSpace(cfg.DefaultProvider) == "" || cfg.DefaultProvider == "none" {
		cfg.DefaultProvider = provider
	}

	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("Saved API key for provider '%s' in %s\n", provider, path)
	return 0
}

func (a *App) runDefault(args []string) int {
	if len(args) < 1 {
		errorStyle.Println("Usage: aifiler default \"model\"")
		return 2
	}
	model := strings.TrimSpace(strings.Join(args, " "))
	if model == "" {
		errorStyle.Println("model cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	cfg.DefaultModel = model
	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("Default model set to '%s' in %s\n", model, path)
	return 0
}

func (a *App) runReset(args []string) int {
	if len(args) < 2 {
		errorStyle.Println("Usage: aifiler reset \"provider\" \"api key\"")
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	apiKey := strings.TrimSpace(args[1])
	if provider == "" || apiKey == "" {
		errorStyle.Println("provider and api key cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	current := strings.TrimSpace(cfg.APIKeys[provider])
	if current == "" {
		warnStyle.Printf("No API key found for provider '%s'\n", provider)
		return 0
	}
	if apiKey != "*" && apiKey != current {
		errorStyle.Printf("Provided api key does not match the stored key for provider '%s'\n", provider)
		return 1
	}

	cfg.APIKeys[provider] = ""
	if cfg.DefaultProvider == provider {
		cfg.DefaultProvider = "none"
	}

	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("API key reset for provider '%s' in %s\n", provider, path)
	return 0
}
