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

var geminiHTTPClient = &http.Client{Timeout: 30 * time.Second}

// GeminiClient calls the Google Gemini REST API directly.
type GeminiClient struct {
	Model  string
	APIKey string
}

func (c *GeminiClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	response, err := c.Prompt(ctx, buildFilenameSuggestionPrompt(originalName, contextHint))
	if err != nil {
		return "", err
	}
	s := normalizeSuggestion(response)
	if s == "" {
		return "", fmt.Errorf("gemini returned empty suggestion")
	}
	return s, nil
}

func (c *GeminiClient) Prompt(ctx context.Context, prompt string) (string, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "gemini-2.0-flash"
	}
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return "", fmt.Errorf("missing API key for Gemini")
	}

	body := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := geminiHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("gemini request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}

	return strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text), nil
}

func (c *GeminiClient) ListModels(ctx context.Context) ([]string, error) {
	apiKey := strings.TrimSpace(c.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for Gemini")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := geminiHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("gemini models request failed %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Models []struct {
			Name             string   `json:"name"`
			SupportedMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range out.Models {
		name := strings.TrimPrefix(m.Name, "models/")
		for _, method := range m.SupportedMethods {
			if method == "generateContent" || method == "generateText" {
				models = append(models, name)
				break
			}
		}
	}
	return models, nil
}
