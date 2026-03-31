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
)

var openaiHTTPClient = &http.Client{Timeout: 30 * time.Second}

// OpenAIClient calls the OpenAI Chat Completions API directly.
type OpenAIClient struct {
	Model  string
	APIKey string
}

func (c *OpenAIClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	response, err := c.Prompt(ctx, buildFilenameSuggestionPrompt(originalName, contextHint))
	if err != nil {
		return "", err
	}
	s := normalizeSuggestion(response)
	if s == "" {
		return "", fmt.Errorf("openai returned empty suggestion")
	}
	return s, nil
}

func (c *OpenAIClient) Prompt(ctx context.Context, prompt string) (string, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "gpt-4o-mini"
	}
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return "", fmt.Errorf("missing API key for OpenAI")
	}

	body := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "user", "content": prompt},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := openaiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("openai request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}

	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func (c *OpenAIClient) ListModels(ctx context.Context) ([]string, error) {
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for OpenAI")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := openaiHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("openai models request failed %d: %s", resp.StatusCode, string(raw))
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
		if strings.HasPrefix(m.ID, "gpt-") || strings.HasPrefix(m.ID, "o1-") || strings.HasPrefix(m.ID, "o3-") {
			models = append(models, m.ID)
		}
	}
	return models, nil
}
