package llm

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/config"
)

type Client interface {
	SuggestName(ctx context.Context, originalName string, contextHint string) (string, error)
	Prompt(ctx context.Context, prompt string) (string, error)
}

type ClientOptions struct {
	Provider string
	Model    string
	Config   config.Config
}

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
	default:
		return &DeterministicClient{}
	}
}

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
	return "Provider is set to 'none'. Configure a real provider with: aifiler set \"provider\" \"api-key\"", nil
}
