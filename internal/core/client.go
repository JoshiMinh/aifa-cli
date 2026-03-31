package core

import (
	"context"
	"fmt"
	"strings"
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
	Config   Config
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
