package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aifiler/internal/core"
)

// runHistory prints recent AI operations to stdout.
func (a *App) runHistory() int {
	path := core.GetHistoryPath() // I will export this in core
	data, err := os.ReadFile(path)
	if err != nil {
		core.WarnStyle.Printf("%s No history found.\n", core.WarnIcon)
		return 0
	}
	var history []core.HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		core.WarnStyle.Printf("%s History is empty.\n", core.WarnIcon)
		return 0
	}

	core.HeaderStyle.Println("Recent AI Operations:")
	for i, entry := range history {
		summary := fmt.Sprintf("%d operations", len(entry.Plan.Operations))
		fmt.Printf("[%d] %s: %s\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"), summary)
	}
	fmt.Println()
	return 0
}

// runUndo reverts the most recent plan from history.
func (a *App) runUndo() int {
	cwd, _ := os.Getwd()
	path := core.GetHistoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		core.ErrorStyle.Printf("%s No history found to undo.\n", core.ErrorIcon)
		return 1
	}
	var history []core.HistoryEntry
	json.Unmarshal(data, &history)

	if len(history) == 0 {
		core.WarnStyle.Printf("%s History is empty.\n", core.WarnIcon)
		return 0
	}

	last := history[len(history)-1]
	core.HeaderStyle.Println("Undoing last operation...")
	for i := len(last.Plan.Operations) - 1; i >= 0; i-- {
		op := last.Plan.Operations[i]
		typ := strings.ToLower(strings.TrimSpace(op.Type))

		switch typ {
		case "create_file", "touch":
			target, _ := core.ResolvePath(cwd, op.Path)
			os.Remove(target)
			core.SuccessStyle.Printf("%s Removed created file: %s\n", core.SuccessIcon, op.Path)
		case "create_dir", "mkdir":
			target, _ := core.ResolvePath(cwd, op.Path)
			os.RemoveAll(target)
			core.SuccessStyle.Printf("%s Removed created dir: %s\n", core.SuccessIcon, op.Path)
		case "rename", "move":
			from, _ := core.ResolvePath(cwd, op.From)
			to, _ := core.ResolvePath(cwd, op.To)
			os.Rename(to, from)
			core.SuccessStyle.Printf("%s Reverted rename: %s -> %s\n", core.SuccessIcon, op.To, op.From)
		case "update_file", "write_file":
			if last.BackupDir != "" {
				backupTarget := filepath.Join(last.BackupDir, op.Path)
				target, _ := core.ResolvePath(cwd, op.Path)
				data, err := os.ReadFile(backupTarget)
				if err == nil {
					os.WriteFile(target, data, 0o644)
					core.SuccessStyle.Printf("%s Restored updated file: %s\n", core.SuccessIcon, op.Path)
				}
			}
		}
	}

	history = history[:len(history)-1]
	newData, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(path, newData, 0o644)

	core.SuccessStyle.Printf("\n%s Undo complete.\n", core.SparkleIcon)
	return 0
}
