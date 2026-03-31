package core

import "github.com/fatih/color"

// Provider holds metadata for a single AI provider.
type Provider struct {
	// Key is the lowercase config key used in config files and API factories.
	Key string
	// DisplayName is the properly capitalised human-readable name.
	DisplayName string
	// RequiresAPIKey indicates whether this provider needs an API key to function.
	RequiresAPIKey bool
	// Style is the terminal color used whenever this provider's name is printed.
	Style *color.Color
}

// Providers is the canonical ordered list of all supported providers.
// Order: direct AI labs first (OpenAI, Anthropic, Gemini), then local (Ollama),
// then gateways (Vercel) last.
var Providers = []Provider{
	{
		Key:            "openai",
		DisplayName:    "OpenAI",
		RequiresAPIKey: true,
		Style:          color.New(color.FgHiWhite, color.Bold),
	},
	{
		Key:            "anthropic",
		DisplayName:    "Anthropic",
		RequiresAPIKey: true,
		Style:          color.RGB(212, 164, 128).Add(color.Bold),
	},
	{
		Key:            "gemini",
		DisplayName:    "Gemini",
		RequiresAPIKey: true,
		Style:          color.New(color.FgHiBlue, color.Bold),
	},
	{
		Key:            "ollama",
		DisplayName:    "Ollama",
		RequiresAPIKey: false,
		Style:          color.New(color.FgHiYellow, color.Bold),
	},
	{
		Key:            "vercel",
		DisplayName:    "Vercel AI Gateway",
		RequiresAPIKey: true,
		Style:          color.New(color.FgHiBlack, color.Bold),
	},
}

// ProviderDisplayNames returns an ordered slice of display names for use in
// interactive prompts.
func ProviderDisplayNames() []string {
	names := make([]string, len(Providers))
	for i, p := range Providers {
		names[i] = p.DisplayName
	}
	return names
}

// ProviderByKey looks up a Provider by its config key (case-insensitive).
// Returns the Provider and true if found, or a zero value and false otherwise.
func ProviderByKey(key string) (Provider, bool) {
	for _, p := range Providers {
		if p.Key == key {
			return p, true
		}
	}
	return Provider{}, false
}

// ProviderByDisplayName looks up a Provider by its display name.
func ProviderByDisplayName(name string) (Provider, bool) {
	for _, p := range Providers {
		if p.DisplayName == name {
			return p, true
		}
	}
	return Provider{}, false
}

// ProviderLabel returns the bold display name for a provider key, falling back
// to the raw key if the provider is not registered.
func ProviderLabel(key string) string {
	if p, ok := ProviderByKey(key); ok {
		if p.Style != nil {
			return p.Style.Sprint(p.DisplayName)
		}
		return p.DisplayName
	}
	return key
}
