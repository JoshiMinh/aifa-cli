package cmds

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/core"
)

func (a *App) runDynamicPrompt(ctx context.Context, prompt string) int {
	currentPrompt := strings.TrimSpace(prompt)
	if currentPrompt == "" {
		a.printHelp()
		return 0
	}

	client, provider, model, err := a.newClient("", "")
	if err != nil {
		core.ErrorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	for {
		workspaceContext := core.BuildWorkspaceContext(a.maxDepth, a.showAll)
		thinking := core.StartThinking("AI is thinking")

		finalPrompt := currentPrompt
		if strings.HasPrefix(currentPrompt, "/") {
			parts := strings.SplitN(currentPrompt[1:], " ", 2)
			intent := parts[0]
			rest := ""
			if len(parts) > 1 {
				rest = parts[1]
			}
			finalPrompt = fmt.Sprintf("FORCE OPERATION TYPE: %s\nUSER REQUEST: %s", strings.ToUpper(intent), rest)
		}

		response, err := client.Prompt(ctx, buildDynamicPrompt(finalPrompt, workspaceContext))
		thinking.Stop("AI response ready")
		if err != nil {
			core.ErrorStyle.Printf("model request failed: %v\n", err)
			return 1
		}

		plan, parseErr := core.ParsePlan(response)
		if parseErr != nil {
			coerceThinking := core.StartThinking("AI is restructuring response as plan")
			coerced, coerceErr := client.Prompt(ctx, buildPlanCoercionPrompt(currentPrompt, response))
			coerceThinking.Stop("Plan conversion ready")
			if coerceErr == nil {
				if repairedPlan, repairedErr := core.ParsePlan(coerced); repairedErr == nil {
					plan = repairedPlan
					parseErr = nil
				}
			}
		}

		core.MutedStyle.Printf("provider=%s model=%s\n", provider, model)
		if parseErr == nil && len(plan.Operations) > 0 {
			result := core.ApplyPlanWithApproval(plan)
			if strings.TrimSpace(result.NextPrompt) == "" {
				return result.ExitCode
			}
			currentPrompt = strings.TrimSpace(result.NextPrompt)
			continue
		}
		if parseErr == nil && len(plan.Operations) == 0 {
			core.WarnStyle.Println("No operations proposed for this prompt.")
			fmt.Println("Try a more specific prompt or request a concrete file/folder change.")
			return 0
		}

		fmt.Println(response)
		return 0
	}
}

func buildDynamicPrompt(userPrompt, workspaceContext string) string {
	return fmt.Sprintf(`You are operating in a local workspace.
If the user request requires filesystem or command actions, return STRICT JSON only in this format:
{"summary":"brief explanation of plan","operations":[{"type":"create_dir|create_file|update_file|rename|delete|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
If the request is informational only, return a normal text response.
Rules for action plans:
- infer file/folder targets from workspace context; do not ask user to describe structure
- paths must be relative and within current directory
- use update_file when modifying existing files
- use run_command only when necessary and keep commands non-interactive
- no markdown fences when returning JSON
Workspace context:
%s
User request: %s`, workspaceContext, userPrompt)
}

func buildPlanCoercionPrompt(userPrompt, modelResponse string) string {
	return fmt.Sprintf(`Convert the following into STRICT JSON only in this exact format:
{"summary":"brief explanation of plan","operations":[{"type":"create_dir|create_file|update_file|rename|delete|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
Rules:
- no explanation text
- no markdown fences
- paths must be relative
User request: %s
Previous response to convert:
%s`, userPrompt, modelResponse)
}
