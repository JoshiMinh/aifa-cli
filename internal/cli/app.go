// Package cli provides the command-line interface logic for the aifiler application.
package cli

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/config"
	"aifiler/internal/llm"
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
	case "history":
		return a.runHistory()
	case "undo":
		return a.runUndo()
	case "search":
		return a.runSearch(args[1:])
	case "ls":
		return a.runLs(args[1:])
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
	headerStyle.Println("aifiler — Vercel AI-Powered File Assistant")
	fmt.Println("Fast, local-first file and folder assistant optimized for Vercel AI Gateway.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aifiler <command> [options]")
	fmt.Println("  aifiler \"<prompt>\"")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create \"<prompt>\"         Create/update files/folders from AI suggestions")
	fmt.Println("  rename \"<prompt>\"         Rename files/folders from AI suggestions")
	fmt.Println("  search -name \"foo\"        Search files by name, ext, or content")
	fmt.Println("  history                   View recent AI operations")
	fmt.Println("  undo                      Revert the last applied AI plan")
	fmt.Println("  ls [-r, -ra]              List workspace files with columns")
	fmt.Println("  list                      Interactive model selection (Vercel recommended)")
	fmt.Println("  set \"provider\"            Save API key for provider (securely prompted)")
	fmt.Println("  default \"model\"           Set default model")
	fmt.Println("  reset \"provider\"          Remove provider API key")
	fmt.Println("  doctor                    Show runtime diagnostics")
	fmt.Println("  help                      Show this help")
	fmt.Println()
	fmt.Println(sparkleIcon + " Vercel AI Gateway Quick Setup (Highly Recommended):")
	fmt.Println("  1. aifiler set \"vercel\"")
	fmt.Println("  2. aifiler default \"openai/gpt-4o-mini\"")
	fmt.Println("  3. aifiler list")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  aifiler create \"create src and README\"")
	fmt.Println("  aifiler \"summarize how to organize this repo\"")
}

func (a *App) newClient(providerOverride, modelOverride string) (llm.Client, string, string, error) {
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return nil, "", "", fmt.Errorf("%s failed to load config: %w\n  %s Tip: Check permissions or run 'aifiler doctor'", errorIcon, err, infoIcon)
	}

	provider := strings.TrimSpace(providerOverride)
	if provider == "" {
		provider = strings.TrimSpace(cfg.DefaultProvider)
	}
	if provider == "" {
		provider = "vercel"
	}

	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = strings.TrimSpace(cfg.DefaultModel)
	}
	if model == "" && provider == "vercel" {
		model = "openai/gpt-4o-mini"
	}

	client := llm.NewClient(llm.ClientOptions{
		Provider: provider,
		Model:    model,
		Config:   cfg,
	})
	return client, provider, model, nil
}
