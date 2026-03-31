package core

import "testing"

func TestNormalizeSuggestion(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"hello world", "hello-world"},
		{"`my_file`", "my-file"},
		{"  spaced  ", "spaced"},
		{"CamelCase", "camelcase"},
	}
	for _, tc := range cases {
		got := NormalizeSuggestion(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeSuggestion(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
