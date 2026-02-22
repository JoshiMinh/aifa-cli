package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ollamaBaseURL = "http://127.0.0.1:11434"

type OllamaClient struct {
	Model string
}

func (c *OllamaClient) SuggestName(ctx context.Context, originalName string, contextHint string) (string, error) {
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "llama3.2"
	}

	prompt := fmt.Sprintf("Suggest a concise kebab-case filename stem for: %q. Context: %s. Return only the filename stem.", originalName, contextHint)
	body := map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaBaseURL+"/api/generate", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 20 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama request failed with status %d", resp.StatusCode)
	}

	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	s := strings.TrimSpace(out.Response)
	s = strings.Trim(s, "\"`")
	s = strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(s), "-"), "_", "-"))
	if s == "" {
		return "", fmt.Errorf("ollama returned empty suggestion")
	}
	return s, nil
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func DetectOllamaModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ollamaBaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: 4 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama tags failed with status %d", resp.StatusCode)
	}

	var result ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(result.Models))
	for _, item := range result.Models {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		models = append(models, item.Name)
	}
	return models, nil
}
