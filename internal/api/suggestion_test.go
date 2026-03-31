package api

import "testing"

func TestNormalizeSuggestion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"  Hello_World  ", "hello-world"},
		{"`my-file_name`", "my-file-name"},
		{`"some text"`, "some-text"},
		{"   ", ""},
		{"already-kebab-case", "already-kebab-case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSuggestion(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeSuggestion(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
