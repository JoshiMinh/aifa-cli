package llm

import (
	"fmt"
	"strings"
)

func buildFilenameSuggestionPrompt(originalName string, contextHint string) string {
	return fmt.Sprintf("Suggest a concise kebab-case filename stem for: %q. Context: %s. Return only the filename stem.", originalName, contextHint)
}

func normalizeSuggestion(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.Trim(value, "\"`")
	value = strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(value), "-"), "_", "-"))
	return value
}
