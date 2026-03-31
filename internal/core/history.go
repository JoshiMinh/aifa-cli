package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryEntry represents a single recorded operation with its plan and backup location.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Plan      AIPlan    `json:"plan"`
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

// SaveStateBeforePlan backs up files that will be modified by the plan.
func SaveStateBeforePlan(cwd string, plan AIPlan) (string, error) {
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
			target, err := ResolvePath(cwd, path)
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

// AppendHistory appends a new entry to the history file (capped at 50 entries).
func AppendHistory(entry HistoryEntry) {
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

// RunHistory prints recent AI operations to stdout.
func RunHistory() int {
	path := getHistoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		WarnStyle.Printf("%s No history found.\n", WarnIcon)
		return 0
	}
	var history []HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		WarnStyle.Printf("%s History is empty.\n", WarnIcon)
		return 0
	}

	HeaderStyle.Println("Recent AI Operations:")
	for i, entry := range history {
		summary := fmt.Sprintf("%d operations", len(entry.Plan.Operations))
		fmt.Printf("[%d] %s: %s\n", i+1, entry.Timestamp.Format(time.RFC3339), summary)
	}
	return 0
}

// RunUndo reverts the most recent plan from history.
func RunUndo() int {
	cwd, _ := os.Getwd()
	path := getHistoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		ErrorStyle.Printf("%s No history found to undo.\n", ErrorIcon)
		return 1
	}
	var history []HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		WarnStyle.Printf("%s History is empty.\n", WarnIcon)
		return 0
	}

	last := history[len(history)-1]
	HeaderStyle.Println("Undoing last operation...")
	for i := len(last.Plan.Operations) - 1; i >= 0; i-- {
		op := last.Plan.Operations[i]
		typ := strings.ToLower(strings.TrimSpace(op.Type))

		switch typ {
		case "create_file", "touch":
			target, _ := ResolvePath(cwd, op.Path)
			os.Remove(target)
			SuccessStyle.Printf("%s Removed created file: %s\n", SuccessIcon, op.Path)
		case "create_dir", "mkdir":
			target, _ := ResolvePath(cwd, op.Path)
			os.RemoveAll(target)
			SuccessStyle.Printf("%s Removed created dir: %s\n", SuccessIcon, op.Path)
		case "rename", "move":
			from, _ := ResolvePath(cwd, op.From)
			to, _ := ResolvePath(cwd, op.To)
			os.Rename(to, from)
			SuccessStyle.Printf("%s Reverted rename: %s -> %s\n", SuccessIcon, op.To, op.From)
		case "update_file", "write_file":
			if last.BackupDir != "" {
				backupTarget := filepath.Join(last.BackupDir, op.Path)
				target, _ := ResolvePath(cwd, op.Path)
				data, err := os.ReadFile(backupTarget)
				if err == nil {
					os.WriteFile(target, data, 0o644)
					SuccessStyle.Printf("%s Restored updated file: %s\n", SuccessIcon, op.Path)
				}
			}
		}
	}

	history = history[:len(history)-1]
	newData, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(path, newData, 0o644)

	SuccessStyle.Printf("\n%s Undo complete.\n", SparkleIcon)
	return 0
}
