package cmds

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"aifiler/internal/api"
	"aifiler/internal/core"

	"github.com/manifoldco/promptui"
)

func (a *App) runList(ctx context.Context) int {
	cfg, _ := core.LoadOrDefault()

	var activeProviders []string
	if strings.TrimSpace(cfg.APIKeys["vercel"]) != "" {
		activeProviders = append(activeProviders, "vercel")
	}
	if strings.TrimSpace(cfg.APIKeys["openai"]) != "" {
		activeProviders = append(activeProviders, "openai")
	}
	if strings.TrimSpace(cfg.APIKeys["anthropic"]) != "" {
		activeProviders = append(activeProviders, "anthropic")
	}
	if val, ok := cfg.APIKeys["gemini"]; ok && strings.TrimSpace(val) != "" {
		activeProviders = append(activeProviders, "gemini")
	} else if val, ok := cfg.APIKeys["google"]; ok && strings.TrimSpace(val) != "" {
		activeProviders = append(activeProviders, "gemini")
	}
	activeProviders = append(activeProviders, "ollama")

	core.HeaderStyle.Println("Active API Providers")
	for _, p := range activeProviders {
		if cfg.DefaultProvider == p {
			fmt.Printf("%s %s (default)\n", core.SuccessIcon, p)
		} else {
			fmt.Printf("  %s\n", p)
		}
	}
	fmt.Println()
	fmt.Printf("default_model: %s\n\n", cfg.DefaultModel)

	activeProvider := cfg.DefaultProvider
	if activeProvider == "" || activeProvider == "none" {
		activeProvider = activeProviders[0]
	}

	if len(activeProviders) > 1 {
		prompt := promptui.Select{
			Label: "Select active provider to configure",
			Items: activeProviders,
			Size:  len(activeProviders),
		}
		_, result, err := prompt.Run()
		if err == nil {
			activeProvider = result
		} else {
			return 0
		}
	}

	core.HeaderStyle.Printf("Fetching dynamic models for '%s'...\n", activeProvider)

	clientInst := api.NewClient(api.ClientOptions{
		Provider: activeProvider,
		Config:   cfg,
	})

	fetched, errModels := clientInst.ListModels(ctx)
	if errModels != nil {
		core.ErrorStyle.Printf("%s Failed to fetch models: %v\n", core.ErrorIcon, errModels)
		return 1
	}

	if len(fetched) > 0 {
		sort.Strings(fetched)
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select default model for %s (Arrow Keys and Enter)", activeProvider),
			Items: fetched,
			Size:  15,
		}
		_, result, err := prompt.Run()
		if err == nil {
			cfg.DefaultModel = result
			cfg.DefaultProvider = activeProvider
			path, errSave := core.Save(cfg)
			if errSave == nil {
				core.SuccessStyle.Printf("%s Default provider set to '%s' and model to '%s' in %s\n", core.SuccessIcon, activeProvider, result, path)
			}
		}
	} else {
		core.WarnStyle.Printf("%s No models found for provider '%s'.\n", core.WarnIcon, activeProvider)
	}

	return 0
}

func (a *App) runSet(args []string) int {
	if len(args) < 1 {
		core.ErrorStyle.Printf("%s Usage: aifiler set \"provider\"\n", core.ErrorIcon)
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	if provider == "" {
		core.ErrorStyle.Println("provider cannot be empty")
		return 2
	}

	var apiKey string
	if len(args) >= 2 {
		core.WarnStyle.Printf("%s Warning: Passing API keys via CLI arguments is insecure. Consider using 'aifiler set \"%s\"' to enter it securely.\n", core.WarnIcon, provider)
		apiKey = strings.TrimSpace(args[1])
	} else {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Enter API Key for %s", provider),
			Mask:  '*',
		}
		res, err := prompt.Run()
		if err != nil {
			core.ErrorStyle.Printf("%s Failed to read API key\n", core.ErrorIcon)
			return 1
		}
		apiKey = strings.TrimSpace(res)
	}

	if apiKey == "" {
		core.ErrorStyle.Println("api key cannot be empty")
		return 2
	}

	cfg, _ := core.LoadOrDefault()
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	cfg.APIKeys[provider] = apiKey
	cfg.DefaultProvider = provider

	path, err := core.Save(cfg)
	if err != nil {
		core.ErrorStyle.Printf("%s failed to save config: %v\n", core.ErrorIcon, err)
		return 1
	}
	core.SuccessStyle.Printf("%s Saved API key for provider '%s' and set as active in %s\n", core.SuccessIcon, provider, path)
	return 0
}

func (a *App) runDefault(args []string) int {
	if len(args) < 1 {
		core.ErrorStyle.Printf("%s Usage: aifiler default \"model\"\n", core.ErrorIcon)
		return 2
	}
	model := strings.TrimSpace(strings.Join(args, " "))
	if model == "" {
		core.ErrorStyle.Printf("%s model cannot be empty\n", core.ErrorIcon)
		return 2
	}

	cfg, _ := core.LoadOrDefault()
	cfg.DefaultModel = model
	path, err := core.Save(cfg)
	if err != nil {
		core.ErrorStyle.Printf("%s failed to save config: %v\n", core.ErrorIcon, err)
		return 1
	}
	core.SuccessStyle.Printf("%s Default model set to '%s' in %s\n", core.SuccessIcon, model, path)
	return 0
}

func (a *App) runReset(args []string) int {
	if len(args) < 1 {
		core.ErrorStyle.Printf("%s Usage: aifiler reset \"provider\"\n", core.ErrorIcon)
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	if provider == "" {
		core.ErrorStyle.Println("provider cannot be empty")
		return 2
	}

	cfg, _ := core.LoadOrDefault()
	current := strings.TrimSpace(cfg.APIKeys[provider])
	if current == "" {
		core.WarnStyle.Printf("%s No API key found for provider '%s'\n", core.WarnIcon, provider)
		return 0
	}

	cfg.APIKeys[provider] = ""
	if cfg.DefaultProvider == provider {
		cfg.DefaultProvider = "none"
	}

	path, err := core.Save(cfg)
	if err != nil {
		core.ErrorStyle.Printf("%s failed to save config: %v\n", core.ErrorIcon, err)
		return 1
	}
	core.SuccessStyle.Printf("%s API key reset for provider '%s' in %s\n", core.SuccessIcon, provider, path)
	return 0
}
