package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	fmt.Println("  create \"<prompt>\"         Create files/folders from AI suggestions")
	fmt.Println("  rename \"<prompt>\"         Rename files/folders from AI suggestions")
	fmt.Println("  list                      List providers, models, and API key status")
	fmt.Println("  set \"provider\" \"api key\"  Save API key for provider")
	fmt.Println("  default \"model\"           Set default model")
	fmt.Println("  reset \"provider\" \"api key\" Remove provider API key")
	fmt.Println("  doctor                    Show runtime diagnostics")
	fmt.Println("  help                      Show this help")
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
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		errorStyle.Println("Usage: aifiler create \"<prompt>\"")
		return 2
	}

	client, _, _, err := a.newClient("", "")
	if err != nil {
		errorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	response, err := client.Prompt(ctx, buildCreatePrompt(prompt))
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

	return applyCreatePlan(plan)
}

func (a *App) runRenameFromPrompt(ctx context.Context, args []string) int {
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		errorStyle.Println("Usage: aifiler rename \"<prompt>\"")
		return 2
	}

	client, _, _, err := a.newClient("", "")
	if err != nil {
		errorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	response, err := client.Prompt(ctx, buildRenamePrompt(prompt))
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

	return applyRenamePlan(plan)
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
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		a.printHelp()
		return 0
	}

	client, provider, model, err := a.newClient("", "")
	if err != nil {
		errorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	response, err := client.Prompt(ctx, prompt)
	if err != nil {
		errorStyle.Printf("model request failed: %v\n", err)
		return 1
	}
	mutedStyle.Printf("provider=%s model=%s\n", provider, model)
	fmt.Println(response)
	return 0
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
}

func buildCreatePrompt(userPrompt string) string {
	return fmt.Sprintf(`You convert requests into filesystem operations.
Return STRICT JSON only in this format:
{"operations":[{"type":"create_dir|create_file","path":"relative/path","content":"optional"}]}
Rules:
- paths must be relative
- no explanation text
- no markdown fences
User request: %s`, userPrompt)
}

func buildRenamePrompt(userPrompt string) string {
	return fmt.Sprintf(`You convert requests into filesystem rename operations.
Return STRICT JSON only in this format:
{"operations":[{"type":"rename","from":"relative/path","to":"relative/path"}]}
Rules:
- paths must be relative
- no explanation text
- no markdown fences
User request: %s`, userPrompt)
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

	var plan aiPlan
	if err := json.Unmarshal([]byte(trimmed), &plan); err != nil {
		return aiPlan{}, err
	}
	return plan, nil
}

func applyCreatePlan(plan aiPlan) int {
	if len(plan.Operations) == 0 {
		warnStyle.Println("No create operations suggested.")
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		errorStyle.Printf("failed to get current folder: %v\n", err)
		return 1
	}

	headerStyle.Println("Applied create operations")
	hadError := false
	for _, op := range plan.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		target, err := resolvePath(cwd, op.Path)
		if err != nil {
			errorStyle.Printf("- %s: %v\n", op.Path, err)
			hadError = true
			continue
		}

		switch typ {
		case "create_dir", "mkdir":
			if err := os.MkdirAll(target, 0o755); err != nil {
				errorStyle.Printf("- create_dir %s: %v\n", op.Path, err)
				hadError = true
				continue
			}
			successStyle.Printf("- created dir: %s\n", op.Path)
		case "create_file", "write_file", "touch":
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
		default:
			warnStyle.Printf("- skipped unknown op type: %s\n", op.Type)
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func applyRenamePlan(plan aiPlan) int {
	if len(plan.Operations) == 0 {
		warnStyle.Println("No rename operations suggested.")
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		errorStyle.Printf("failed to get current folder: %v\n", err)
		return 1
	}

	headerStyle.Println("Applied rename operations")
	hadError := false
	for _, op := range plan.Operations {
		typ := strings.ToLower(strings.TrimSpace(op.Type))
		if typ != "rename" && typ != "move" {
			warnStyle.Printf("- skipped unknown op type: %s\n", op.Type)
			continue
		}

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
	}

	if hadError {
		return 1
	}
	return 0
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
