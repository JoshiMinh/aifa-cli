package cli

import (
	"context"
	"fmt"
	"strings"
)

func (a *App) runDynamicPrompt(ctx context.Context, prompt string) int {
	currentPrompt := strings.TrimSpace(prompt)
	if currentPrompt == "" {
		a.printHelp()
		return 0
	}

	client, provider, model, err := a.newClient("", "")
	if err != nil {
		errorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	for {
		workspaceContext := buildWorkspaceContext(a.maxDepth, a.showAll)
		thinking := startThinking("AI is thinking")

		// Handle / prefixes for intent forcing
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
		thinking.stop("AI response ready")
		if err != nil {
			errorStyle.Printf("model request failed: %v\n", err)
			return 1
		}

		plan, parseErr := parsePlan(response)
		if parseErr != nil {
			coerceThinking := startThinking("AI is restructuring response as plan")
			coerced, coerceErr := client.Prompt(ctx, buildPlanCoercionPrompt(currentPrompt, response))
			coerceThinking.stop("Plan conversion ready")
			if coerceErr == nil {
				if repairedPlan, repairedErr := parsePlan(coerced); repairedErr == nil {
					plan = repairedPlan
					parseErr = nil
				}
			}
		}

		mutedStyle.Printf("provider=%s model=%s\n", provider, model)
		if parseErr == nil && len(plan.Operations) > 0 {
			result := applyPlanWithApproval(plan)
			if strings.TrimSpace(result.NextPrompt) == "" {
				return result.ExitCode
			}
			currentPrompt = strings.TrimSpace(result.NextPrompt)
			continue
		}
		if parseErr == nil && len(plan.Operations) == 0 {
			warnStyle.Println("No operations proposed for this prompt.")
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
