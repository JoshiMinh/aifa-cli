package core

import (
	"encoding/json"
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

// GetHistoryPath returns the absolute path to history.json.
func GetHistoryPath() string {
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
	path := GetHistoryPath()
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
