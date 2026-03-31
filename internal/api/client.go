package api

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/core"
)

// Client defines the interface for AI operations.
type Client interface {
	SuggestName(ctx context.Context, originalName string, contextHint string) (string, error)
	Prompt(ctx context.Context, prompt string) (string, error)
	ListModels(ctx context.Context) ([]string, error)
}

// ClientOptions specifies configuration required to instantiate a new Client.
type ClientOptions struct {
	Provider string
	Model    string
	Config   core.Config
}

// NewClient creates and returns the appropriate Client implementation based on the provider.
func NewClient(opts ClientOptions) Client {
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
		return &DeterministicClient{}
	}
}

// DeterministicClient provides fallback responses when no real AI provider is configured.
type DeterministicClient struct{}

func (c *DeterministicClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	_ = ctx
	name := strings.TrimSpace(originalName)
	if name == "" {
		return "", fmt.Errorf("empty filename")
	}
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.Join(strings.Fields(name), " ")
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name, nil
}

func (c *DeterministicClient) Prompt(ctx context.Context, prompt string) (string, error) {
	_ = ctx
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("empty prompt")
	}
	return "Provider is set to 'none'. Configure a real provider with: aifiler set \"provider\"", nil
}

func (c *DeterministicClient) ListModels(ctx context.Context) ([]string, error) {
	return nil, nil
}
