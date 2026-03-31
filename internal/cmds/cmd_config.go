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

// runList fetches and displays available models for the currently active provider.
func (a *App) runList(ctx context.Context) int {
	cfg, _ := core.LoadOrDefault()

	providerKey := strings.TrimSpace(cfg.DefaultProvider)
	if providerKey == "" || providerKey == "none" {
		core.WarnStyle.Printf("%s No active provider set. Run 'aifiler provider' to configure one.\n", core.WarnIcon)
		return 1
	}

	p, ok := core.ProviderByKey(providerKey)
	if !ok {
		core.ErrorStyle.Printf("%s Unknown provider '%s'. Run 'aifiler provider' to set a valid one.\n", core.ErrorIcon, providerKey)
		return 1
	}

	providerLabel := p.DisplayName
	if p.Style != nil {
		providerLabel = p.Style.Sprint(p.DisplayName)
	}

	core.HeaderStyle.Printf("\n  Fetching models for %s...\n\n", providerLabel)

	clientInst := api.NewClient(core.ClientOptions{
		Provider: p.Key,
		Config:   cfg,
	})

	fetched, err := clientInst.ListModels(ctx)
	if err != nil {
		core.ErrorStyle.Printf("%s Failed to fetch models: %v\n", core.ErrorIcon, err)
		return 1
	}

	if len(fetched) == 0 {
		core.WarnStyle.Printf("%s No models found for %s.\n", core.WarnIcon, providerLabel)
		return 0
	}

	sort.Strings(fetched)
	core.HeaderStyle.Printf("  Available models for %s:\n\n", providerLabel)
	for _, m := range fetched {
		fmt.Printf("    %s\n", core.MutedStyle.Sprint(m))
	}
	fmt.Println()
	return 0
}

// runProvider is the primary interactive configuration command.
// It lets the user switch providers, set API keys, and browse models.
func (a *App) runProvider() int {
	cfg, _ := core.LoadOrDefault()

	activeKey := strings.TrimSpace(cfg.DefaultProvider)
	labels := make([]string, len(core.Providers))
	for i, p := range core.Providers {
		label := p.DisplayName
		if p.Key == activeKey {
			label += " (active)"
		}
		labels[i] = label
	}

	selectPrompt := promptui.Select{
		Label: "Select provider to configure",
		Items: labels,
		Size:  len(labels),
	}
	idx, _, err := selectPrompt.Run()
	if err != nil {
		return 0
	}
	chosen := core.Providers[idx]

	// Ask what to do with the chosen provider.
	actions := []string{"Set as active provider", "Set API key", "Set active + update API key", "Clear API key", "Cancel"}
	actionPrompt := promptui.Select{
		Label: fmt.Sprintf("Action for %s", chosen.DisplayName),
		Items: actions,
		Size:  len(actions),
	}
	actionIdx, _, err := actionPrompt.Run()
	if err != nil || actionIdx == len(actions)-1 {
		return 0
	}

	switch actionIdx {
	case 0: // Set as active only
		cfg.DefaultProvider = chosen.Key
		path, saveErr := core.Save(cfg)
		if saveErr != nil {
			core.ErrorStyle.Printf("%s Failed to save config: %v\n", core.ErrorIcon, saveErr)
			return 1
		}
		core.SuccessStyle.Printf("%s Active provider set to '%s' in %s\n", core.SuccessIcon, chosen.DisplayName, path)

	case 1: // Set API key only
		if !chosen.RequiresAPIKey {
			core.WarnStyle.Printf("%s %s does not require an API key.\n", core.WarnIcon, chosen.DisplayName)
			return 0
		}
		apiKey, ok := promptAPIKey(chosen.DisplayName)
		if !ok {
			return 1
		}
		if cfg.APIKeys == nil {
			cfg.APIKeys = map[string]string{}
		}
		cfg.APIKeys[chosen.Key] = apiKey
		path, saveErr := core.Save(cfg)
		if saveErr != nil {
			core.ErrorStyle.Printf("%s Failed to save config: %v\n", core.ErrorIcon, saveErr)
			return 1
		}
		core.SuccessStyle.Printf("%s API key saved for '%s' in %s\n", core.SuccessIcon, chosen.DisplayName, path)

	case 2: // Set active + update API key
		cfg.DefaultProvider = chosen.Key
		if chosen.RequiresAPIKey {
			apiKey, ok := promptAPIKey(chosen.DisplayName)
			if !ok {
				return 1
			}
			if cfg.APIKeys == nil {
				cfg.APIKeys = map[string]string{}
			}
			cfg.APIKeys[chosen.Key] = apiKey
		}
		path, saveErr := core.Save(cfg)
		if saveErr != nil {
			core.ErrorStyle.Printf("%s Failed to save config: %v\n", core.ErrorIcon, saveErr)
			return 1
		}
		core.SuccessStyle.Printf("%s Active provider set to '%s' in %s\n", core.SuccessIcon, chosen.DisplayName, path)

	case 3: // Clear API key
		if cfg.APIKeys == nil || cfg.APIKeys[chosen.Key] == "" {
			core.WarnStyle.Printf("%s No API key found for '%s'.\n", core.WarnIcon, chosen.DisplayName)
			return 0
		}
		cfg.APIKeys[chosen.Key] = ""
		if cfg.DefaultProvider == chosen.Key {
			cfg.DefaultProvider = "none"
		}
		path, saveErr := core.Save(cfg)
		if saveErr != nil {
			core.ErrorStyle.Printf("%s Failed to save config: %v\n", core.ErrorIcon, saveErr)
			return 1
		}
		core.SuccessStyle.Printf("%s API key cleared for '%s' in %s\n", core.SuccessIcon, chosen.DisplayName, path)
	}

	return 0
}

// promptAPIKey interactively asks the user for an API key for the named provider.
// Returns (key, true) on success or ("", false) on failure.
func promptAPIKey(providerName string) (string, bool) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("API key for %s", providerName),
		Mask:  '*',
	}
	res, err := prompt.Run()
	if err != nil {
		core.ErrorStyle.Printf("%s Failed to read API key\n", core.ErrorIcon)
		return "", false
	}
	key := strings.TrimSpace(res)
	if key == "" {
		core.ErrorStyle.Println("API key cannot be empty")
		return "", false
	}
	return key, true
}
