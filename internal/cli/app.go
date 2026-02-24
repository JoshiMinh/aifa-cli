package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"aifiler/internal/config"
	"aifiler/internal/llm"
	"aifiler/internal/models"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.printHelp()
		return 0
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "help", "-h", "--help":
		a.printHelp()
		return 0
	case "doctor":
		return a.runDoctor()
	case "list":
		return a.runList(ctx)
	case "set":
		return a.runSet(args[1:])
	case "default":
		return a.runDefault(args[1:])
	case "reset":
		return a.runReset(args[1:])
	case "create":
		return a.runCreate(ctx, args[1:])
	case "rename":
		return a.runRenameFromPrompt(ctx, args[1:])
	default:
		return a.runDynamicPrompt(ctx, strings.Join(args, " "))
	}
}

func (a *App) printHelp() {
	headerStyle.Println("aifiler — AI File Assistant")
	fmt.Println("AI-powered, local-first file and folder assistant.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aifiler <command> [options]")
	fmt.Println("  aifiler \"<prompt>\"")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create \"<prompt>\"         Create/update files/folders from AI suggestions")
	fmt.Println("  rename \"<prompt>\"         Rename files/folders from AI suggestions")
	fmt.Println("  list                      List providers, models, and API key status")
	fmt.Println("  set \"provider\" \"api key\"  Save API key for provider")
	fmt.Println("  default \"model\"           Set default model")
	fmt.Println("  reset \"provider\" \"api key\" Remove provider API key")
	fmt.Println("  doctor                    Show runtime diagnostics")
	fmt.Println("  help                      Show this help")
	fmt.Println()
	fmt.Println("Behavior:")
	fmt.Println("  - Prompts automatically include current workspace structure")
	fmt.Println("  - Any file/folder/command action requires approval before execution")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  aifiler create \"create src and README\"")
	fmt.Println("  aifiler rename \"rename docs to documentation folder\"")
	fmt.Println("  aifiler doctor")
	fmt.Println("  aifiler \"summarize how to organize this repo\"")
	fmt.Println()
	fmt.Println("Vercel quick setup:")
	fmt.Println("  aifiler set \"vercel\" \"<ai-gateway-api-key>\"")
	fmt.Println("  aifiler default \"openai/gpt-4o-mini\"")
	fmt.Println("  aifiler list")
}

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

func (a *App) runRenameFromPrompt(ctx context.Context, args []string) int {
	currentPrompt := strings.TrimSpace(strings.Join(args, " "))
	if currentPrompt == "" {
		errorStyle.Println("Usage: aifiler rename \"<prompt>\"")
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
		response, err := client.Prompt(ctx, buildRenamePrompt(currentPrompt, workspaceContext))
		thinking.stop("AI response ready")
		if err != nil {
			errorStyle.Printf("rename failed: %v\n", err)
			return 1
		}

		plan, err := parsePlan(response)
		if err != nil {
			warnStyle.Println("Could not parse structured rename plan; model response:")
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

func (a *App) runList(ctx context.Context) int {
	registry, err := models.LoadDefaultRegistry()
	if err != nil {
		errorStyle.Printf("failed to load model registry: %v\n", err)
		return 1
	}
	cfg, _ := config.LoadOrDefault()

	headerStyle.Println("Available providers and models")
	registry.Print("")

	fmt.Println()
	headerStyle.Println("Configured API keys")
	providers := make([]string, 0, len(registry.Providers)+len(cfg.APIKeys))
	providerSet := map[string]struct{}{}
	for provider := range registry.Providers {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" || provider == "none" {
			continue
		}
		providerSet[provider] = struct{}{}
	}
	for provider := range cfg.APIKeys {
		provider = strings.ToLower(strings.TrimSpace(provider))
		if provider == "" {
			continue
		}
		providerSet[provider] = struct{}{}
	}
	for provider := range providerSet {
		providers = append(providers, provider)
	}
	sort.Strings(providers)

	if len(providers) == 0 {
		fmt.Println("- none")
	} else {
		for _, provider := range providers {
			status := "not-set"
			if strings.TrimSpace(cfg.APIKeys[provider]) != "" {
				status = "set"
			}
			fmt.Printf("- %s: %s\n", provider, status)
		}
	}

	fmt.Println()
	fmt.Printf("default_provider: %s\n", cfg.DefaultProvider)
	fmt.Printf("default_model: %s\n", cfg.DefaultModel)

	ollamaModels, err := llm.DetectOllamaModels(ctx)
	if err == nil && len(ollamaModels) > 0 {
		fmt.Println()
		headerStyle.Println("Detected local Ollama models")
		for _, model := range ollamaModels {
			fmt.Printf("  - %s\n", model)
		}
	}

	vercelModels, err := llm.DetectVercelModels(ctx, cfg.APIKeys["vercel"], "")
	if err == nil && len(vercelModels) > 0 {
		fmt.Println()
		headerStyle.Println("Detected Vercel AI Gateway models")
		for _, model := range vercelModels {
			fmt.Printf("  - %s\n", model)
		}
	}
	return 0
}

func (a *App) runDoctor() int {
	headerStyle.Println("aifiler diagnostics")

	if cwd, err := os.Getwd(); err == nil {
		fmt.Printf("cwd: %s\n", cwd)
	}
	if exePath, err := os.Executable(); err == nil {
		fmt.Printf("executable: %s\n", exePath)
	}

	configured := strings.TrimSpace(os.Getenv(models.RegistryPathEnvVar))
	if configured == "" {
		fmt.Printf("%s: (not set)\n", models.RegistryPathEnvVar)
	} else {
		fmt.Printf("%s: %s\n", models.RegistryPathEnvVar, configured)
	}

	resolved, err := models.ResolveRegistryPath(models.DefaultRegistryPath)
	if err != nil {
		errorStyle.Printf("registry: unresolved (%v)\n", err)
		return 1
	}

	successStyle.Printf("registry: %s\n", resolved)
	return 0
}

func (a *App) runSet(args []string) int {
	if len(args) < 2 {
		errorStyle.Println("Usage: aifiler set \"provider\" \"api key\"")
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	apiKey := strings.TrimSpace(args[1])
	if provider == "" || apiKey == "" {
		errorStyle.Println("provider and api key cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	if cfg.APIKeys == nil {
		cfg.APIKeys = map[string]string{}
	}
	cfg.APIKeys[provider] = apiKey
	if strings.TrimSpace(cfg.DefaultProvider) == "" || cfg.DefaultProvider == "none" {
		cfg.DefaultProvider = provider
	}

	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("Saved API key for provider '%s' in %s\n", provider, path)
	return 0
}

func (a *App) runDefault(args []string) int {
	if len(args) < 1 {
		errorStyle.Println("Usage: aifiler default \"model\"")
		return 2
	}
	model := strings.TrimSpace(strings.Join(args, " "))
	if model == "" {
		errorStyle.Println("model cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	cfg.DefaultModel = model
	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("Default model set to '%s' in %s\n", model, path)
	return 0
}

func (a *App) runReset(args []string) int {
	if len(args) < 2 {
		errorStyle.Println("Usage: aifiler reset \"provider\" \"api key\"")
		return 2
	}
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	apiKey := strings.TrimSpace(args[1])
	if provider == "" || apiKey == "" {
		errorStyle.Println("provider and api key cannot be empty")
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	current := strings.TrimSpace(cfg.APIKeys[provider])
	if current == "" {
		warnStyle.Printf("No API key found for provider '%s'\n", provider)
		return 0
	}
	if apiKey != "*" && apiKey != current {
		errorStyle.Printf("Provided api key does not match the stored key for provider '%s'\n", provider)
		return 1
	}

	cfg.APIKeys[provider] = ""
	if cfg.DefaultProvider == provider {
		cfg.DefaultProvider = "none"
	}

	path, err := config.Save(cfg)
	if err != nil {
		errorStyle.Printf("failed to save config: %v\n", err)
		return 1
	}
	successStyle.Printf("API key reset for provider '%s' in %s\n", provider, path)
	return 0
}

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
		workspaceContext := buildWorkspaceContext()
		thinking := startThinking("AI is thinking")
		response, err := client.Prompt(ctx, buildDynamicPrompt(currentPrompt, workspaceContext))
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

func (a *App) newClient(providerOverride, modelOverride string) (llm.Client, string, string, error) {
	cfg, _ := config.LoadOrDefault()
	registry, err := models.LoadDefaultRegistry()
	if err != nil {
		return nil, "", "", err
	}

	provider := strings.TrimSpace(providerOverride)
	if provider == "" {
		provider = strings.TrimSpace(cfg.DefaultProvider)
	}
	if provider == "" {
		provider = "none"
	}

	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = strings.TrimSpace(cfg.DefaultModel)
	}
	if model == "" {
		model = registry.DefaultModelForProvider(provider)
	}

	client := llm.NewClient(llm.ClientOptions{
		Provider: provider,
		Model:    model,
		Config:   cfg,
	})
	return client, provider, model, nil
}

type aiPlan struct {
	Operations []aiOperation `json:"operations"`
}

type aiOperation struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	Command string `json:"command"`
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

func buildRenamePrompt(userPrompt, workspaceContext string) string {
	return fmt.Sprintf(`You convert requests into filesystem rename operations.
Return STRICT JSON only in this format:
{"operations":[{"type":"rename|run_command","from":"relative/path","to":"relative/path","command":"optional"}]}
Rules:
- infer file/folder targets from workspace context; do not ask user to describe structure
- paths must be relative and within current directory
- no explanation text
- no markdown fences
Workspace context:
%s
User request: %s`, workspaceContext, userPrompt)
}

func buildDynamicPrompt(userPrompt, workspaceContext string) string {
	return fmt.Sprintf(`You are operating in a local workspace.
If the user request requires filesystem or command actions, return STRICT JSON only in this format:
{"operations":[{"type":"create_dir|create_file|update_file|rename|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
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
{"operations":[{"type":"create_dir|create_file|update_file|rename|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
Rules:
- no explanation text
- no markdown fences
- paths must be relative
User request: %s
Previous response to convert:
%s`, userPrompt, modelResponse)
}

func parsePlan(response string) (aiPlan, error) {
	trimmed := strings.TrimSpace(response)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimPrefix(trimmed, "json")
		trimmed = strings.TrimSpace(trimmed)
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	if !strings.HasPrefix(trimmed, "{") {
		if candidate := extractFirstJSONObject(trimmed); candidate != "" {
			trimmed = candidate
		}
	}

	var plan aiPlan
	if err := json.Unmarshal([]byte(trimmed), &plan); err != nil {
		return aiPlan{}, err
	}
	return plan, nil
}

func extractFirstJSONObject(input string) string {
	start := strings.Index(input, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(input); i++ {
		switch input[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(input[start : i+1])
			}
		}
	}
	return ""
}

type planApplyResult struct {
	ExitCode   int
	NextPrompt string
}

func applyPlanWithApproval(plan aiPlan) planApplyResult {
	if len(plan.Operations) == 0 {
		warnStyle.Println("No operations suggested.")
		return planApplyResult{ExitCode: 0}
	}

	cwd, err := os.Getwd()
	if err != nil {
		errorStyle.Printf("failed to get current folder: %v\n", err)
		return planApplyResult{ExitCode: 1}
	}

	headerStyle.Println("Proposed operations")
	for i, op := range plan.Operations {
		renderOperationLine(i+1, op)
	}

	fmt.Println()
	headerStyle.Println("Proposed tree")
	renderProposedTree(plan)

	decision := promptApplyDecision("Apply these operations")
	if decision.Approve == false {
		if strings.TrimSpace(decision.NextPrompt) != "" {
			return planApplyResult{ExitCode: 0, NextPrompt: decision.NextPrompt}
		}
		warnStyle.Println("Plan was not approved. No changes were made.")
		return planApplyResult{ExitCode: 0}
	}

	headerStyle.Println("Applying operations")
	hadError := false
	for _, op := range plan.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))

		switch typ {
		case "create_dir", "mkdir":
			target, err := resolvePath(cwd, op.Path)
			if err != nil {
				errorStyle.Printf("- %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			if err := os.MkdirAll(target, 0o755); err != nil {
				errorStyle.Printf("- create_dir %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			successStyle.Printf("- created dir: %s\n", op.Path)
		case "create_file", "touch":
			target, err := resolvePath(cwd, op.Path)
			if err != nil {
				errorStyle.Printf("- %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				errorStyle.Printf("- prepare dir for %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			if _, err := os.Stat(target); err == nil {
				warnStyle.Printf("- skipped existing file: %s\n", op.Path)
				continue
			}
			if err := os.WriteFile(target, []byte(op.Content), 0o644); err != nil {
				errorStyle.Printf("- create_file %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			successStyle.Printf("- created file: %s\n", op.Path)
		case "update_file", "write_file":
			target, err := resolvePath(cwd, op.Path)
			if err != nil {
				errorStyle.Printf("- %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				errorStyle.Printf("- prepare dir for %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			if err := os.WriteFile(target, []byte(op.Content), 0o644); err != nil {
				errorStyle.Printf("- update_file %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			successStyle.Printf("- updated file: %s\n", op.Path)
		case "rename", "move":
			fromPath := strings.TrimSpace(op.From)
			if fromPath == "" {
				fromPath = strings.TrimSpace(op.Path)
			}
			toPath := strings.TrimSpace(op.To)
			from, err := resolvePath(cwd, fromPath)
			if err != nil {
				errorStyle.Printf("- invalid from path '%s': %v\n", fromPath, err)
				hadError = true
				continue
			}
			to, err := resolvePath(cwd, toPath)
			if err != nil {
				errorStyle.Printf("- invalid to path '%s': %v\n", toPath, err)
				hadError = true
				continue
			}

			if _, err := os.Stat(from); err != nil {
				errorStyle.Printf("- source missing: %s\n", fromPath)
				hadError = true
				continue
			}
			if _, err := os.Stat(to); err == nil {
				errorStyle.Printf("- target exists: %s\n", toPath)
				hadError = true
				continue
			}
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				errorStyle.Printf("- create target dir for %s: %v\n", toPath, err)
				hadError = true
				continue
			}
			if err := os.Rename(from, to); err != nil {
				errorStyle.Printf("- rename %s -> %s failed: %v\n", fromPath, toPath, err)
				hadError = true
				continue
			}
			successStyle.Printf("- renamed: %s -> %s\n", fromPath, toPath)
		case "run_command":
			command := strings.TrimSpace(op.Command)
			if command == "" {
				warnStyle.Println("- skipped empty command")
				continue
			}
			if !confirmApproval(fmt.Sprintf("Run command '%s'", command)) {
				warnStyle.Printf("- skipped command: %s\n", command)
				continue
			}
			if err := runCommand(cwd, command); err != nil {
				errorStyle.Printf("- command failed: %s (%v)\n", command, err)
				hadError = true
				continue
			}
			successStyle.Printf("- command succeeded: %s\n", command)
		default:
			warnStyle.Printf("- skipped unknown op type: %s\n", op.Type)
		}
	}

	if hadError {
		return planApplyResult{ExitCode: 1}
	}
	return planApplyResult{ExitCode: 0}
}

type applyDecision struct {
	Approve    bool
	NextPrompt string
}

func promptApplyDecision(message string) applyDecision {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s? [y/N or type next prompt]: ", message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return applyDecision{}
	}
	choice := strings.TrimSpace(input)
	lowerChoice := strings.ToLower(choice)
	if lowerChoice == "y" || lowerChoice == "yes" {
		return applyDecision{Approve: true}
	}
	if lowerChoice == "n" || lowerChoice == "no" || lowerChoice == "" {
		return applyDecision{}
	}
	return applyDecision{NextPrompt: choice}
}

func confirmApproval(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s? [y/N]: ", message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	choice := strings.ToLower(strings.TrimSpace(input))
	return choice == "y" || choice == "yes"
}

func runCommand(cwd, command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-lc", command)
	}
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(strings.TrimSpace(string(output)))
	}
	return err
}

func buildWorkspaceContext() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "Current directory unavailable"
	}

	snapshot := collectWorkspaceSnapshot(cwd, 4, 300)
	return fmt.Sprintf("Current directory: %s\nWorkspace files/folders:\n%s", cwd, snapshot)
}

func collectWorkspaceSnapshot(root string, maxDepth, maxEntries int) string {
	entries := []string{}
	count := 0

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root {
			return nil
		}

		base := strings.ToLower(d.Name())
		if d.IsDir() {
			switch base {
			case ".git", "node_modules", ".idea", ".vscode":
				return filepath.SkipDir
			}
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		if rel == "." {
			return nil
		}

		depth := strings.Count(rel, "/") + 1
		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if count >= maxEntries {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			rel += "/"
		}
		entries = append(entries, rel)
		count++
		return nil
	})

	if len(entries) == 0 {
		return "(workspace appears empty)"
	}

	if len(entries) >= maxEntries {
		entries = append(entries, "... (truncated)")
	}

	return strings.Join(entries, "\n")
}

func resolvePath(cwd, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty path")
	}
	target := value
	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	target = filepath.Clean(target)
	rel, err := filepath.Rel(cwd, target)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes current directory")
	}
	return target, nil
}

type thinkingIndicator struct {
	message string
	done    chan struct{}
	once    sync.Once
}

func startThinking(message string) *thinkingIndicator {
	indicator := &thinkingIndicator{
		message: message,
		done:    make(chan struct{}),
	}

	go func() {
		frames := []string{"|", "/", "-", "\\"}
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		index := 0
		for {
			select {
			case <-indicator.done:
				return
			case <-ticker.C:
				fmt.Printf("\r%s", thinkingStyle.Sprintf("%s %s", frames[index%len(frames)], indicator.message))
				index++
			}
		}
	}()

	return indicator
}

func (i *thinkingIndicator) stop(message string) {
	i.once.Do(func() {
		close(i.done)
		fmt.Printf("\r%s\r", strings.Repeat(" ", 120))
		if strings.TrimSpace(message) != "" {
			infoStyle.Println(message)
		}
	})
}

type proposalTreeNode struct {
	name        string
	children    map[string]*proposalTreeNode
	annotations []string
}

func newProposalTreeNode(name string) *proposalTreeNode {
	return &proposalTreeNode{
		name:     name,
		children: map[string]*proposalTreeNode{},
	}
}

func renderOperationLine(index int, op aiOperation) {
	typ := strings.ToLower(strings.TrimSpace(op.Type))
	prefix := fmt.Sprintf("%d.", index)

	switch typ {
	case "create_dir", "mkdir":
		fmt.Printf("%s %s %s\n", mutedStyle.Sprint(prefix), opCreateStyle.Sprint(typ), pathStyle.Sprint(strings.TrimSpace(op.Path)))
	case "create_file", "touch":
		fmt.Printf("%s %s %s\n", mutedStyle.Sprint(prefix), opCreateStyle.Sprint(typ), pathStyle.Sprint(strings.TrimSpace(op.Path)))
	case "update_file", "write_file":
		fmt.Printf("%s %s %s\n", mutedStyle.Sprint(prefix), opUpdateStyle.Sprint(typ), pathStyle.Sprint(strings.TrimSpace(op.Path)))
	case "rename", "move":
		fmt.Printf("%s %s %s %s %s\n", mutedStyle.Sprint(prefix), opRenameStyle.Sprint(typ), pathStyle.Sprint(strings.TrimSpace(op.From)), mutedStyle.Sprint("->"), pathStyle.Sprint(strings.TrimSpace(op.To)))
	case "run_command":
		fmt.Printf("%s %s %s\n", mutedStyle.Sprint(prefix), opCommandStyle.Sprint("run_command"), commandStyle.Sprint(strings.TrimSpace(op.Command)))
	default:
		fmt.Printf("%s %s\n", mutedStyle.Sprint(prefix), warnStyle.Sprint(op.Type))
	}
}

func renderProposedTree(plan aiPlan) {
	root := newProposalTreeNode(".")

	for _, op := range plan.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		switch typ {
		case "create_dir", "mkdir":
			addProposalPath(root, strings.TrimSpace(op.Path), "create_dir")
		case "create_file", "touch":
			addProposalPath(root, strings.TrimSpace(op.Path), "create_file")
		case "update_file", "write_file":
			addProposalPath(root, strings.TrimSpace(op.Path), "update_file")
		case "rename", "move":
			addProposalPath(root, strings.TrimSpace(op.From), "rename_from")
			addProposalPath(root, strings.TrimSpace(op.To), "rename_to")
		}
	}

	if len(root.children) == 0 {
		warnStyle.Println("(no path-based changes to render)")
		return
	}

	printProposalTree(root, "")
}

func addProposalPath(root *proposalTreeNode, path, annotation string) {
	cleaned := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if cleaned == "" || cleaned == "." {
		return
	}
	parts := strings.Split(cleaned, "/")
	node := root
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." {
			continue
		}
		child, ok := node.children[part]
		if !ok {
			child = newProposalTreeNode(part)
			node.children[part] = child
		}
		node = child
	}
	if annotation != "" {
		node.annotations = append(node.annotations, annotation)
	}
}

func printProposalTree(node *proposalTreeNode, prefix string) {
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}
	sort.Strings(names)

	for index, name := range names {
		child := node.children[name]
		isLast := index == len(names)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if isLast {
			branch = "└── "
			nextPrefix = prefix + "    "
		}

		fmt.Printf("%s%s%s", treeBranchStyle.Sprint(prefix), treeBranchStyle.Sprint(branch), pathStyle.Sprint(child.name))
		if len(child.annotations) > 0 {
			fmt.Printf(" %s", formatAnnotations(child.annotations))
		}
		fmt.Println()
		printProposalTree(child, nextPrefix)
	}
}

func formatAnnotations(annotations []string) string {
	seen := map[string]struct{}{}
	ordered := make([]string, 0, len(annotations))
	for _, item := range annotations {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		ordered = append(ordered, key)
	}

	parts := make([]string, 0, len(ordered))
	for _, key := range ordered {
		switch key {
		case "create_dir":
			parts = append(parts, opCreateStyle.Sprint("[create_dir]"))
		case "create_file":
			parts = append(parts, opCreateStyle.Sprint("[create_file]"))
		case "update_file":
			parts = append(parts, opUpdateStyle.Sprint("[update_file]"))
		case "rename_from":
			parts = append(parts, opRenameStyle.Sprint("[rename_from]"))
		case "rename_to":
			parts = append(parts, opRenameStyle.Sprint("[rename_to]"))
		default:
			parts = append(parts, mutedStyle.Sprint("["+key+"]"))
		}
	}
	return strings.Join(parts, " ")
}
