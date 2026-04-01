package cmds

import (
	"encoding/json"
	"fmt"
	"os"

	"aifiler/internal/core"
)

// runHistory prints recent AI operations to stdout.
func (a *App) runHistory() int {
	path := core.GetHistoryPath()
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

	messages, err := core.RevertPlan(cwd, last)
	if err != nil {
		core.ErrorStyle.Printf("%s Undo failed: %v\n", core.ErrorIcon, err)
		return 1
	}

	for _, msg := range messages {
		core.SuccessStyle.Printf("%s %s\n", core.SuccessIcon, msg)
	}

	core.RemoveLastHistory()

	core.SuccessStyle.Printf("\n%s Undo complete.\n", core.SparkleIcon)
	return 0
}
