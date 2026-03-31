package core

import "testing"

func TestTraversal(t *testing.T) {
	ctx := BuildWorkspaceContext(1, false)
	if ctx == "" {
		t.Error("Expected non-empty workspace context")
	}
}
