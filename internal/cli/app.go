// Package cli provides the command-line interface logic for the aifiler application.
package cli

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/config"
	"aifiler/internal/llm"
	"aifiler/internal/models"
)

// App represents the main CLI application.
type App struct{}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Run executes the CLI application with the given arguments.
// It returns the exit code.
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

func (a *App) newClient(providerOverride, modelOverride string) (llm.Client, string, string, error) {
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load config: %w", err)
	}

	registry, err := models.LoadDefaultRegistry()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load registry: %w", err)
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
