package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AIPlan represents the structured plan returned by the LLM.
type AIPlan struct {
	Summary    string      `json:"summary"`
	NextPrompt string      `json:"next_prompt"`
	Operations []Operation `json:"operations"`
}

// Operation represents a single filesystem or shell operation in a plan.
type Operation struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	Command string `json:"command"`
}

// ApplyResult is returned after user approves or rejects a plan.
type ApplyResult struct {
	ExitCode   int
	NextPrompt string
}

// ParsePlan attempts to parse a raw JSON string into an AIPlan.
func ParsePlan(raw string) (AIPlan, error) {
	var p AIPlan
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	if err := json.Unmarshal([]byte(cleaned), &p); err != nil {
		return p, err
	}
	return p, nil
}


func ExecuteOperation(cwd string, op Operation) error {
	typ := strings.ToLower(strings.TrimSpace(op.Type))
	switch typ {
	case "create_dir", "mkdir":
		target, err := ResolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		return os.MkdirAll(target, 0o755)
	case "create_file", "touch":
		target, err := ResolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(target), 0o755)
		return os.WriteFile(target, []byte(op.Content), 0o644)
	case "update_file", "write_file":
		target, err := ResolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(target), 0o755)
		return os.WriteFile(target, []byte(op.Content), 0o644)
	case "rename", "move":
		from, err := ResolvePath(cwd, op.From)
		if err != nil {
			return err
		}
		to, err := ResolvePath(cwd, op.To)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(to), 0o755)
		return os.Rename(from, to)
	case "delete", "remove":
		target, err := ResolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		return os.RemoveAll(target)
	case "run_command":
		cmdArgs := strings.Fields(op.Command)
		if len(cmdArgs) == 0 {
			return nil
		}
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = cwd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unknown operation type: %s", typ)
	}
}

// ResolvePath resolves a relative path safely within cwd.
func ResolvePath(cwd, path string) (string, error) {
	abs := filepath.Join(cwd, path)
	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes current directory: %s", path)
	}
	return abs, nil
}
