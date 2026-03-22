package cli

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) runCreate(ctx context.Context, args []string) int {
	currentPrompt := strings.TrimSpace(strings.Join(args, " "))
	if currentPrompt == "" {
		errorStyle.Println("Usage: aifiler create \"<prompt>\"")
		return 2
	}

	client, _, _, err := a.newClient("", "")
	if err != nil {
		errorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	for {
		workspaceContext := buildWorkspaceContext()
		thinking := startThinking("AI is thinking")
		response, err := client.Prompt(ctx, buildCreatePrompt(currentPrompt, workspaceContext))
		thinking.stop("AI response ready")
		if err != nil {
			errorStyle.Printf("create failed: %v\n", err)
			return 1
		}

		plan, err := parsePlan(response)
		if err != nil {
			warnStyle.Println("Could not parse structured create plan; model response:")
			fmt.Println(response)
			return 0
		}

		result := applyPlanWithApproval(plan)
		if strings.TrimSpace(result.NextPrompt) == "" {
			return result.ExitCode
		}
		currentPrompt = strings.TrimSpace(result.NextPrompt)
	}
}

func buildCreatePrompt(userPrompt, workspaceContext string) string {
	return fmt.Sprintf(`You convert requests into filesystem operations.
Return STRICT JSON only in this format:
{"operations":[{"type":"create_dir|create_file|update_file|rename|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
Rules:
- infer file/folder targets from workspace context; do not ask user to describe structure
- paths must be relative and within current directory
- use update_file when modifying an existing file
- use run_command only when necessary and keep commands non-interactive
- no explanation text
- no markdown fences
Workspace context:
%s
User request: %s`, workspaceContext, userPrompt)
}
