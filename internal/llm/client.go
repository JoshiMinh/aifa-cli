package llm

import (
	"context"
	"fmt"
	"strings"

	"aifa/internal/config"
)

type Client interface {
	SuggestName(ctx context.Context, originalName string, contextHint string) (string, error)
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
