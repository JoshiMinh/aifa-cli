package core

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.DefaultProvider != "none" {
		t.Errorf("Expected DefaultProvider 'none', got %q", cfg.DefaultProvider)
	}

	if cfg.APIKeys == nil {
		t.Fatal("Expected APIKeys map to be initialized")
	}

	expectedKeys := []string{"openai", "anthropic", "gemini", "vercel", "ollama"}
	for _, key := range expectedKeys {
		if _, exists := cfg.APIKeys[key]; !exists {
			t.Errorf("Expected pre-populated key %q in APIKeys map", key)
		}
	}
}
