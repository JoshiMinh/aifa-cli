package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"aifiler/internal/core"
	"github.com/schollz/progressbar/v3"
)

// ApplyPlanWithApproval shows the plan to the user, prompts for approval, and executes.
func ApplyPlanWithApproval(p core.AIPlan) core.ApplyResult {
	cwd, _ := os.Getwd()

	core.HeaderStyle.Println("\nPlan Summary")
	fmt.Printf("  %s\n\n", p.Summary)

	core.HeaderStyle.Println("Proposed Operations")
	for i, op := range p.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		desc := ""
		switch typ {
		case "create_dir", "mkdir":
			desc = fmt.Sprintf("%s %s", core.FolderIcon, op.Path)
		case "create_file", "touch":
			desc = fmt.Sprintf("%s %s", core.FileIcon, op.Path)
		case "update_file", "write_file":
			desc = fmt.Sprintf("%s %s (modified)", core.EditIcon, op.Path)
		case "rename", "move":
			desc = fmt.Sprintf("%s %s -> %s", core.RenameIcon, op.From, op.To)
		case "delete", "remove":
			desc = fmt.Sprintf("%s %s (deleted)", core.DeleteIcon, op.Path)
		case "run_command":
			desc = fmt.Sprintf("%s %s", core.CommandIcon, op.Command)
		}
		fmt.Printf("  %d. %s\n", i+1, desc)
	}

	if p.NextPrompt != "" {
		fmt.Printf("\n  %s %s\n", core.InfoIcon, core.MutedStyle.Sprintf("This plan includes a follow-up: %q", p.NextPrompt))
	}

	fmt.Printf("\nApply these operations? [y/N or type next prompt]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "y" || input == "yes" {
		backupDir, _ := core.SaveStateBeforePlan(cwd, p)

		bar := progressbar.Default(int64(len(p.Operations)), "Applying changes")
		for _, op := range p.Operations {
			if err := core.ExecuteOperation(cwd, op); err != nil {
				bar.Exit()
				core.ErrorStyle.Printf("\n%s Operation failed: %v\n", core.ErrorIcon, err)
				return core.ApplyResult{ExitCode: 1}
			}
			bar.Add(1)
		}
		fmt.Println()

		core.AppendHistory(core.HistoryEntry{
			Timestamp: time.Now(),
			Plan:      p,
			BackupDir: backupDir,
		})

		core.SuccessStyle.Printf("%s Operations applied successfully.\n", core.SuccessIcon)
		if p.NextPrompt != "" {
			return core.ApplyResult{ExitCode: 0, NextPrompt: p.NextPrompt}
		}
		return core.ApplyResult{ExitCode: 0}
	} else if input != "" && input != "n" && input != "no" {
		return core.ApplyResult{ExitCode: 0, NextPrompt: input}
	}

	fmt.Println("Plan was not approved. No changes were made.")
	return core.ApplyResult{ExitCode: 0}
}
