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

var selectTemplates = &promptui.SelectTemplates{
	FuncMap: promptui.FuncMap,
	Label:   "{{ . }}",
	Active:  "\u27a4 {{ . | cyan }}", // fallback for simple strings
	Inactive: "  {{ . }}",
	Selected: "\u2714 {{ . | green }}",
	Details:  "",
	Help:     "",
}

// providerItem is used to pass structured data to the select templates.
type providerItem struct {
	Name     string
	Key      string
	IsActive bool
}

// providerSelectTemplates defines how provider items are rendered.
var providerSelectTemplates = &promptui.SelectTemplates{
	FuncMap: promptui.FuncMap,
	Label:   "{{ . }}",
	Active:   "\u27a4 {{ .Name | cyan }} key: {{ .Key | faint }}{{ if .IsActive }} (active){{ end }}",
	Inactive: "  {{ .Name }} key: {{ .Key | faint }}{{ if .IsActive }} (active){{ end }}",
	Selected: "\u2714 {{ .Name | green }}",
	Details:  "",
	Help:     "",
}

func init() {
	// Initialize custom primary color if needed, but 'cyan' is built-in and matches PrimaryColor.
	// For better compatibility with Windows and to avoid the nesting bug, 
	// we keep the templates clean.
}

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

	selectPrompt := promptui.Select{
		Label:     "Select default model",
		Items:     fetched,
		Size:      10,
		Templates: selectTemplates,
		HideHelp:  true,
	}

	_, selected, err := selectPrompt.Run()
	if err != nil {
		return 0
	}

	cfg.DefaultModel = selected
	path, saveErr := core.Save(cfg)
	if saveErr != nil {
		core.ErrorStyle.Printf("%s Failed to save config: %v\n", core.ErrorIcon, saveErr)
		return 1
	}

	core.SuccessStyle.Printf("%s Default model set to '%s' in %s\n", core.SuccessIcon, selected, path)
	return 0
}

// runProvider is the primary interactive configuration command.
// It lets the user switch providers, set API keys, and browse models.
func (a *App) runProvider() int {
	cfg, _ := core.LoadOrDefault()
	activeKey := strings.TrimSpace(cfg.DefaultProvider)

	items := make([]providerItem, len(core.Providers))
	for i, p := range core.Providers {
		items[i] = providerItem{
			Name:     p.DisplayName,
			Key:      p.Key,
			IsActive: p.Key == activeKey,
		}
	}

	selectPrompt := promptui.Select{
		Label:     "Select provider to configure",
		Items:     items,
		Size:      len(items),
		Templates: providerSelectTemplates,
		HideHelp:  true,
	}
	idx, _, err := selectPrompt.Run()
	if err != nil {
		return 0
	}
	chosen := core.Providers[idx]

	// Ask what to do with the chosen provider.
	actions := []string{"Set as active provider", "Set API key", "Set active + update API key", "Clear API key", "Cancel"}
	actionPrompt := promptui.Select{
		Label:     fmt.Sprintf("Action for %s", chosen.DisplayName),
		Items:     actions,
		Size:      len(actions),
		Templates: selectTemplates,
		HideHelp:  true,
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
