package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"aifiler/internal/config"
	"aifiler/internal/llm"

	"github.com/manifoldco/promptui"
)

func (a *App) runList(ctx context.Context) int {
	cfg, _ := config.LoadOrDefault()
	
	allowedProviders := map[string]bool{
		"google": true, "alibaba": true, "anthropic": true, "deepseek": true, "meta": true, "openai": true, "xai": true,
	}

	var activeProviders []string
	for provider, key := range cfg.APIKeys {
		if strings.TrimSpace(key) != "" {
			activeProviders = append(activeProviders, provider)
		}
	}

	headerStyle.Println("Configured API Providers")
	if len(activeProviders) == 0 {
		warnStyle.Printf("%s No providers configured.\n  %s Tip: Start with Vercel AI Gateway for the best experience.\n  Run: aifiler set \"vercel\"\n", warnIcon, infoIcon)
	} else {
		if len(activeProviders) > 1 {
			warnStyle.Printf("%s Multiple API providers found! (%s)\n  %s Security Tip: For efficiency and safety, only one API provider should be active at once.\n  Please use 'aifiler reset' to remove unused keys.\n\n", warnIcon, strings.Join(activeProviders, ", "), infoIcon)
		}
		for _, provider := range activeProviders {
			fmt.Printf("%s active: %s\n", successIcon, provider)
		}
	}

	fmt.Println()
	fmt.Printf("default_model: %s\n", cfg.DefaultModel)
	fmt.Println()

	var allowedModels []string
	activeProvider := ""
	if len(activeProviders) > 0 {
		activeProvider = activeProviders[0] // just pick the first one for detection
		if cfg.DefaultProvider != "" && cfg.DefaultProvider != "none" {
			activeProvider = cfg.DefaultProvider
		}
	}

	if activeProvider != "" {
		headerStyle.Printf("Fetching dynamic models for '%s'...\n", activeProvider)
		var fetched []string
		var errModels error
		
		switch activeProvider {
		case "vercel":
			fetched, errModels = llm.DetectVercelModels(ctx, cfg.APIKeys["vercel"], "")
		case "ollama":
			fetched, errModels = llm.DetectOllamaModels(ctx)
		default:
			// For others, if we don't have detection, we can't show much without a registry.
			warnStyle.Printf("%s Live model detection not yet implemented for '%s'.\n", infoIcon, activeProvider)
		}

		if errModels == nil && len(fetched) > 0 {
			for _, m := range fetched {
				// Vercel models usually are provider/name
				if activeProvider == "vercel" {
					parts := strings.Split(m, "/")
					if len(parts) > 0 && allowedProviders[strings.ToLower(parts[0])] {
						allowedModels = append(allowedModels, m)
					}
				} else {
					allowedModels = append(allowedModels, m)
				}
			}
			sort.Strings(allowedModels)
		}
	}

	if len(allowedModels) > 0 {
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select default model for %s (Arrow Keys and Enter)", activeProvider),
			Items: allowedModels,
			Size:  15,
		}
		_, result, err := prompt.Run()
		if err == nil {
			cfg.DefaultModel = result
			cfg.DefaultProvider = activeProvider
			path, errSave := config.Save(cfg)
			if errSave == nil {
				successStyle.Printf("%s Default model set to '%s' in %s\n", successIcon, result, path)
			}
		}
	} else if activeProvider == "vercel" {
		warnStyle.Printf("%s No approved text models were found on your Vercel AI Gateway.\n  %s Tip: Ensure your Gateway has 'openai', 'anthropic', or 'google' providers connected.\n", warnIcon, infoIcon)
	}

	return 0
}

func (a *App) runSet(args []string) int {
	if len(args) < 1 {
		errorStyle.Printf("%s Usage: aifiler set \"provider\"\n", errorIcon)
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	if provider == "" {
		errorStyle.Println("provider cannot be empty")
		return 2
	}

	var apiKey string
	if len(args) >= 2 {
		warnStyle.Printf("%s Warning: Passing API keys via CLI arguments is insecure. Consider using 'aifiler set \"%s\"' to enter it securely.\n", warnIcon, provider)
		apiKey = strings.TrimSpace(args[1])
	} else {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Enter API Key for %s", provider),
			Mask:  '*',
		}
		res, err := prompt.Run()
		if err != nil {
			errorStyle.Printf("%s Failed to read API key\n", errorIcon)
			return 1
		}
		apiKey = strings.TrimSpace(res)
	}

	if apiKey == "" {
		errorStyle.Println("api key cannot be empty")
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
		errorStyle.Printf("%s failed to save config: %v\n", errorIcon, err)
		return 1
	}
	successStyle.Printf("%s Saved API key for provider '%s' in %s\n", successIcon, provider, path)
	return 0
}

func (a *App) runDefault(args []string) int {
	if len(args) < 1 {
		errorStyle.Printf("%s Usage: aifiler default \"model\"\n", errorIcon)
		return 2
	}
	model := strings.TrimSpace(strings.Join(args, " "))
	if model == "" {
		errorStyle.Printf("%s model cannot be empty\n", errorIcon)
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	cfg.DefaultModel = model
	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("%s failed to save config: %v\n", errorIcon, err)
		return 1
	}
	successStyle.Printf("%s Default model set to '%s' in %s\n", successIcon, model, path)
	return 0
}

func (a *App) runReset(args []string) int {
	if len(args) < 1 {
		errorStyle.Printf("%s Usage: aifiler reset \"provider\"\n", errorIcon)
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	if provider == "" {
		errorStyle.Println("provider cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	current := strings.TrimSpace(cfg.APIKeys[provider])
	if current == "" {
		warnStyle.Printf("%s No API key found for provider '%s'\n", warnIcon, provider)
		return 0
	}

	cfg.APIKeys[provider] = ""
	if cfg.DefaultProvider == provider {
		cfg.DefaultProvider = "none"
	}

	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("%s failed to save config: %v\n", errorIcon, err)
		return 1
	}
	successStyle.Printf("%s API key reset for provider '%s' in %s\n", successIcon, provider, path)
	return 0
}
