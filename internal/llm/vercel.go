package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const vercelGatewayBaseURL = "https://ai-gateway.vercel.sh/v1"

var vercelPromptHTTPClient = &http.Client{Timeout: 30 * time.Second}
var vercelModelsHTTPClient = &http.Client{Timeout: 12 * time.Second}

type VercelGatewayClient struct {
	Model   string
	APIKey  string
	BaseURL string
}

type vercelModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *VercelGatewayClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	response, err := c.Prompt(ctx, buildFilenameSuggestionPrompt(originalName, contextHint))
	if err != nil {
		return "", err
	}
	s := normalizeSuggestion(response)
	if s == "" {
		return "", fmt.Errorf("vercel gateway returned empty suggestion")
	}
	return s, nil
}

func (c *VercelGatewayClient) Prompt(ctx context.Context, prompt string) (string, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	apiKey, baseURL, err := resolveVercelGatewayConfig(c.APIKey, c.BaseURL)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := vercelPromptHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("vercel gateway request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	if len(out.Choices) == 0 {
		return "", fmt.Errorf("vercel gateway returned no choices")
	}

	content, err := extractChatContent(out.Choices[0].Message.Content)
	if err != nil {
		return "", err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("vercel gateway returned empty content")
	}

	return content, nil
}

func DetectVercelModels(ctx context.Context, apiKey string, baseURL string) ([]string, error) {
	resolvedAPIKey, resolvedBaseURL, err := resolveVercelGatewayConfig(apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolvedBaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+resolvedAPIKey)

	resp, err := vercelModelsHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("vercel gateway models request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out vercelModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(out.Data))
	for _, item := range out.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		models = append(models, id)
	}
	return models, nil
}

func resolveVercelGatewayConfig(apiKey string, baseURL string) (string, string, error) {
	resolvedAPIKey := strings.TrimSpace(apiKey)
	if resolvedAPIKey == "" {
		resolvedAPIKey = strings.TrimSpace(os.Getenv("AI_GATEWAY_API_KEY"))
	}
	if resolvedAPIKey == "" {
		return "", "", fmt.Errorf("missing API key for provider 'vercel' (use: aifiler set \"vercel\" \"<api-key>\" or set AI_GATEWAY_API_KEY)")
	}

	resolvedBaseURL := strings.TrimSpace(baseURL)
	if resolvedBaseURL == "" {
		resolvedBaseURL = strings.TrimSpace(os.Getenv("AI_GATEWAY_BASE_URL"))
	}
	if resolvedBaseURL == "" {
		resolvedBaseURL = vercelGatewayBaseURL
	}
	return resolvedAPIKey, strings.TrimSuffix(resolvedBaseURL, "/"), nil
}

func extractChatContent(raw any) (string, error) {
	switch value := raw.(type) {
	case string:
		return value, nil
	case []any:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			typeValue, _ := obj["type"].(string)
			if typeValue != "" && typeValue != "text" {
				continue
			}
			text, _ := obj["text"].(string)
			if strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
		if len(parts) == 0 {
			return "", fmt.Errorf("no text content in response")
		}
		return strings.Join(parts, "\n"), nil
	default:
		return "", fmt.Errorf("unsupported content format in response")
	}
}
