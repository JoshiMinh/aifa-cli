package api

import (
	"fmt"
	"strings"
)

// buildFilenameSuggestionPrompt creates a prompt asking the AI to suggest a concise, kebab-case filename.
func buildFilenameSuggestionPrompt(originalName string, contextHint string) string {
	return fmt.Sprintf("Suggest a concise kebab-case filename stem for: %q. Context: %s. Return only the filename stem.", originalName, contextHint)
}

// normalizeSuggestion cleans and formats the AI's response into a valid, kebab-case filename stem.
func normalizeSuggestion(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.Trim(value, "`\"")
	value = strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(value), "-"), "_", "-"))
	return value
}
