package cli

import (
	"testing"
)

func TestTraversal(t *testing.T) {
	// Simple test to ensure traversal logic doesn't crash
	ctx := buildWorkspaceContext(1, false)
	if ctx == "" {
		t.Error("Expected non-empty workspace context")
	}
}
