package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Plan      aiPlan    `json:"plan"`
	BackupDir string    `json:"backup_dir"`
}

func getHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aifiler", "history.json")
}

func getBackupBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aifiler", "backups")
}

func saveStateBeforePlan(cwd string, plan aiPlan) (string, error) {
	ts := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(getBackupBaseDir(), ts)

	hasBackups := false
	for _, op := range plan.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		if typ == "update_file" || typ == "write_file" || typ == "rename" || typ == "move" {
			path := op.Path
			if typ == "rename" || typ == "move" {
				path = op.From
			}
			if strings.TrimSpace(path) == "" {
				continue
			}
			target, err := resolvePath(cwd, path)
			if err == nil {
				data, err := os.ReadFile(target)
				if err == nil {
					os.MkdirAll(backupDir, 0o755)
					backupTarget := filepath.Join(backupDir, path)
					os.MkdirAll(filepath.Dir(backupTarget), 0o755)
					os.WriteFile(backupTarget, data, 0o644)
					hasBackups = true
				}
			}
		}
	}
	if !hasBackups {
		return "", nil
	}
	return backupDir, nil
}

func appendHistory(entry HistoryEntry) {
	path := getHistoryPath()
	os.MkdirAll(filepath.Dir(path), 0o755)

	var history []HistoryEntry
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &history)
	}

	history = append(history, entry)
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	newData, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(path, newData, 0o644)
}

func (a *App) runHistory() int {
	path := getHistoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		warnStyle.Printf("%s No history found.\n", warnIcon)
		return 0
	}
	var history []HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		warnStyle.Printf("%s History is empty.\n", warnIcon)
		return 0
	}

	headerStyle.Println("Recent AI Operations:")
	for i, entry := range history {
		summary := fmt.Sprintf("%d operations", len(entry.Plan.Operations))
		fmt.Printf("[%d] %s: %s\n", i+1, entry.Timestamp.Format(time.RFC3339), summary)
	}
	return 0
}

func (a *App) runUndo() int {
	cwd, _ := os.Getwd()
	path := getHistoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		errorStyle.Printf("%s No history found to undo.\n", errorIcon)
		return 1
	}
	var history []HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		warnStyle.Printf("%s History is empty.\n", warnIcon)
		return 0
	}

	last := history[len(history)-1]

	headerStyle.Println("Undoing last operation...")
	for i := len(last.Plan.Operations) - 1; i >= 0; i-- {
		op := last.Plan.Operations[i]
		typ := strings.ToLower(strings.TrimSpace(op.Type))

		switch typ {
		case "create_file", "touch":
			target, _ := resolvePath(cwd, op.Path)
			os.Remove(target)
			successStyle.Printf("%s Removed created file: %s\n", successIcon, op.Path)
		case "create_dir", "mkdir":
			target, _ := resolvePath(cwd, op.Path)
			os.RemoveAll(target)
			successStyle.Printf("%s Removed created dir: %s\n", successIcon, op.Path)
		case "rename", "move":
			from, _ := resolvePath(cwd, op.From)
			to, _ := resolvePath(cwd, op.To)
			os.Rename(to, from)
			successStyle.Printf("%s Reverted rename: %s -> %s\n", successIcon, op.To, op.From)
		case "update_file", "write_file":
			if last.BackupDir != "" {
				backupTarget := filepath.Join(last.BackupDir, op.Path)
				target, _ := resolvePath(cwd, op.Path)
				data, err := os.ReadFile(backupTarget)
				if err == nil {
					os.WriteFile(target, data, 0o644)
					successStyle.Printf("%s Restored updated file: %s\n", successIcon, op.Path)
				}
			}
		}
	}

	history = history[:len(history)-1]
	newData, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(path, newData, 0o644)

	successStyle.Printf("\n%s Undo complete.\n", sparkleIcon)
	return 0
}
