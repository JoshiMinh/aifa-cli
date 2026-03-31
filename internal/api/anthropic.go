package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aifiler/internal/core"
)

var anthropicHTTPClient = &http.Client{Timeout: 30 * time.Second}

// AnthropicClient calls the Anthropic Messages API directly.
type AnthropicClient struct {
	Model  string
	APIKey string
}

func (c *AnthropicClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	response, err := c.Prompt(ctx, core.BuildFilenameSuggestionPrompt(originalName, contextHint))
	if err != nil {
		return "", err
	}
	s := core.NormalizeSuggestion(response)
	if s == "" {
		return "", fmt.Errorf("anthropic returned empty suggestion")
	}
	return s, nil
}

func (c *AnthropicClient) Prompt(ctx context.Context, prompt string) (string, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "claude-3-7-sonnet-20250219"
	}
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return "", fmt.Errorf("missing API key for Anthropic")
	}

	body := map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to create anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := anthropicHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("anthropic request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("failed to decode anthropic response: %w", err)
	}

	for _, block := range out.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			return strings.TrimSpace(block.Text), nil
		}
	}
	return "", fmt.Errorf("anthropic returned empty content")
}

func (c *AnthropicClient) ListModels(ctx context.Context) ([]string, error) {
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for Anthropic")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := anthropicHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("anthropic models request failed %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range out.Data {
		if strings.HasPrefix(m.ID, "claude-") {
			models = append(models, m.ID)
		}
	}
	return models, nil
}
