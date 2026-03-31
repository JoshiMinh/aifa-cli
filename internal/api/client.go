package api

import (
	"strings"

	"aifiler/internal/core"
)

// NewClient creates and returns the appropriate Client implementation based on the provider.
func NewClient(opts core.ClientOptions) core.Client {
	provider := strings.ToLower(strings.TrimSpace(opts.Provider))
	switch provider {
	case "ollama":
		return &OllamaClient{Model: opts.Model}
	case "vercel":
		apiKey := ""
		if opts.Config.APIKeys != nil {
			apiKey = strings.TrimSpace(opts.Config.APIKeys["vercel"])
		}
		return &VercelGatewayClient{Model: opts.Model, APIKey: apiKey}
	case "gemini", "google":
		apiKey := ""
		if opts.Config.APIKeys != nil {
			if val, ok := opts.Config.APIKeys["gemini"]; ok && val != "" {
				apiKey = val
			} else if val, ok := opts.Config.APIKeys["google"]; ok && val != "" {
				apiKey = val
			}
		}
		return &GeminiClient{Model: opts.Model, APIKey: strings.TrimSpace(apiKey)}
	case "anthropic":
		apiKey := ""
		if opts.Config.APIKeys != nil {
			apiKey = strings.TrimSpace(opts.Config.APIKeys["anthropic"])
		}
		return &AnthropicClient{Model: opts.Model, APIKey: apiKey}
	case "openai":
		apiKey := ""
		if opts.Config.APIKeys != nil {
			apiKey = strings.TrimSpace(opts.Config.APIKeys["openai"])
		}
		return &OpenAIClient{Model: opts.Model, APIKey: apiKey}
	default:
		return &core.DeterministicClient{}
	}
}
