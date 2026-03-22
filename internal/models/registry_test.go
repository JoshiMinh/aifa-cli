package models

import "testing"

func TestDefaultModelForProvider(t *testing.T) {
	reg := Registry{
		Providers: map[string]Provider{
			"vercel": {Default: "gpt-4o-mini", Models: []string{"gpt-4o-mini"}},
			"ollama": {Default: "llama3", Models: []string{"llama3"}},
		},
	}

	if got := reg.DefaultModelForProvider("Vercel"); got != "gpt-4o-mini" {
		t.Errorf("Expected 'gpt-4o-mini', got %q", got)
	}

	if got := reg.DefaultModelForProvider("unknown"); got != "" {
		t.Errorf("Expected empty string for unknown provider, got %q", got)
	}
}
