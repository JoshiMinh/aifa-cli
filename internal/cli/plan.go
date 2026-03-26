package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

type aiPlan struct {
	Summary    string        `json:"summary"`
	NextPrompt string        `json:"next_prompt"`
	Operations []operation   `json:"operations"`
}

type operation struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	Command string `json:"command"`
}

type applyResult struct {
	ExitCode   int
	NextPrompt string
}

func parsePlan(raw string) (aiPlan, error) {
	var p aiPlan
	cleaned := strings.TrimSpace(raw)
	// Remove markdown fences if present
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	if err := json.Unmarshal([]byte(cleaned), &p); err != nil {
		return p, err
	}
	return p, nil
}

func applyPlanWithApproval(p aiPlan) applyResult {
	cwd, _ := os.Getwd()

	headerStyle.Println("\nPlan Summary")
	fmt.Printf("  %s\n\n", p.Summary)

	headerStyle.Println("Proposed Operations")
	for i, op := range p.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		desc := ""
		switch typ {
		case "create_dir", "mkdir":
			desc = fmt.Sprintf("%s %s", folderIcon, op.Path)
		case "create_file", "touch":
			desc = fmt.Sprintf("%s %s", fileIcon, op.Path)
		case "update_file", "write_file":
			desc = fmt.Sprintf("%s %s (modified)", editIcon, op.Path)
		case "rename", "move":
			desc = fmt.Sprintf("%s %s -> %s", renameIcon, op.From, op.To)
		case "delete", "remove":
			desc = fmt.Sprintf("%s %s (deleted)", deleteIcon, op.Path)
		case "run_command":
			desc = fmt.Sprintf("%s %s", commandIcon, op.Command)
		}
		fmt.Printf("  %d. %s\n", i+1, desc)
	}

	if p.NextPrompt != "" {
		fmt.Printf("\n  %s %s\n", infoIcon, mutedStyle.Sprintf("This plan includes a follow-up: %q", p.NextPrompt))
	}

	fmt.Printf("\nApply these operations? [y/N or type next prompt]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "y" || input == "yes" {
		backupDir, _ := saveStateBeforePlan(cwd, p)
		
		bar := progressbar.Default(int64(len(p.Operations)), "Applying changes")
		for _, op := range p.Operations {
			if err := executeOperation(cwd, op); err != nil {
				bar.Exit()
				errorStyle.Printf("\n%s Operation failed: %v\n", errorIcon, err)
				return applyResult{ExitCode: 1}
			}
			bar.Add(1)
		}
		fmt.Println()

		appendHistory(HistoryEntry{
			Timestamp: time.Now(),
			Plan:      p,
			BackupDir: backupDir,
		})

		successStyle.Printf("%s Operations applied successfully.\n", successIcon)
		if p.NextPrompt != "" {
			return applyResult{ExitCode: 0, NextPrompt: p.NextPrompt}
		}
		return applyResult{ExitCode: 0}
	} else if input != "" && input != "n" && input != "no" {
		// Treat any other input as a refinement prompt
		return applyResult{ExitCode: 0, NextPrompt: input}
	}

	fmt.Println("Plan was not approved. No changes were made.")
	return applyResult{ExitCode: 0}
}

func executeOperation(cwd string, op operation) error {
	typ := strings.ToLower(strings.TrimSpace(op.Type))
	switch typ {
	case "create_dir", "mkdir":
		target, err := resolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		return os.MkdirAll(target, 0o755)
	case "create_file", "touch":
		target, err := resolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(target), 0o755)
		return os.WriteFile(target, []byte(op.Content), 0o644)
	case "update_file", "write_file":
		target, err := resolvePath(cwd, op.Path)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(target), 0o755)
		return os.WriteFile(target, []byte(op.Content), 0o644)
	case "rename", "move":
		from, err := resolvePath(cwd, op.From)
		if err != nil {
			return err
		}
		to, err := resolvePath(cwd, op.To)
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Dir(to), 0o755)
		return os.Rename(from, to)
	case "delete", "remove":
		target, err := resolvePath(cwd, op.Path)
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

func resolvePath(cwd, path string) (string, error) {
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
