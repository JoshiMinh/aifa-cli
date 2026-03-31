// Package cmds provides the command-line interface logic for the aifiler application.
package cmds

import (
	"context"
	"fmt"
	"strings"

	"aifiler/internal/api"
	"aifiler/internal/core"
)

// App represents the main CLI application.
type App struct {
	maxDepth int
	showAll  bool
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		maxDepth: 1,
		showAll:  false,
	}
}

// Run executes the CLI application with the given arguments.
func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.printHelp()
		return 0
	}

	var remainingArgs []string
	for _, arg := range args {
		switch arg {
		case "-r":
			a.maxDepth = 2
		case "-ra":
			a.maxDepth = 0
		case "-all":
			a.showAll = true
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	if len(remainingArgs) == 0 {
		a.printHelp()
		return 0
	}

	cmd := strings.ToLower(strings.TrimSpace(remainingArgs[0]))
	switch cmd {
	case "help", "-h", "--help":
		a.printHelp()
		return 0
	case "doctor":
		return a.runDoctor()
	case "list":
		return a.runList(ctx)
	case "provider":
		return a.runProvider()
	case "history":
		return a.runHistory()
	case "undo":
		return a.runUndo()
	default:
		return a.runDynamicPrompt(ctx, strings.Join(remainingArgs, " "))
	}
}

func (a *App) printHelp() {
	fmt.Println()
	core.HeaderStyle.Println("        _ _____ __         ")
	core.HeaderStyle.Println("  ____ _(_) __(_) /__  _____")
	core.HeaderStyle.Println(" / __ `/ / /_/ / / _ \\/ ___/")
	core.HeaderStyle.Println("/ /_/ / / __/ / /  __/ /    ")
	core.HeaderStyle.Println("\\__,_/_/_/ /_/_/\\___/_/     ")
	core.HeaderStyle.Println()

	core.HeaderStyle.Println("  aifiler — AI-Powered Filesystem Assistant")
	fmt.Println("  --------------------------------------------------")
	fmt.Println("  A local-first assistant that translates natural language into safely")
	fmt.Println("  orchestrated filesystem operations.")
	fmt.Println()

	core.HeaderStyle.Println("  MAIN COMMAND")
	fmt.Printf("    %-25s %s\n", core.PathStyle.Sprint("aifiler \"<prompt>\""), "Execute an AI-powered plan based on your request")
	fmt.Printf("    %-25s %s\n", core.PathStyle.Sprint("aifiler \"/<intent> ...\""), "Force a specific operation (e.g., /create, /rename, /delete)")
	fmt.Println()

	core.HeaderStyle.Println("  OPTIONS")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("-r"), "Scan root and immediate subfolders (one-level)")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("-ra"), "Fully recursive scan (all levels)")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("-all"), "Include all file entries in AI context (no truncation)")
	fmt.Println()

	core.HeaderStyle.Println("  PROVIDERS")
	for _, p := range core.Providers {
		name := p.DisplayName
		if p.Style != nil {
			name = p.Style.Sprint(p.DisplayName)
		}
		fmt.Printf("    %-30s key: %s\n", name, core.MutedStyle.Sprint(p.Key))
	}
	fmt.Println()

	core.HeaderStyle.Println("  UTILITIES")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("list"), "List available models for the active provider")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("provider"), "Switch provider, set API keys, browse models")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("history"), "View recent AI operations")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("undo"), "Revert the last applied AI plan")
	fmt.Printf("    %-25s %s\n", core.MutedStyle.Sprint("doctor"), "Show runtime diagnostics")
	fmt.Printf("    %s\n", core.MutedStyle.Sprintf("Config file: %s", core.ConfigPath()))
	fmt.Println()

	core.HeaderStyle.Println("  EXAMPLES")
	fmt.Println("    " + core.MutedStyle.Sprint("aifiler \"create a clean src layout with README\""))
	fmt.Println("    " + core.MutedStyle.Sprint("aifiler -r \"rename all .js files to .ts\""))
	fmt.Println("    " + core.MutedStyle.Sprint("aifiler -ra -all \"search for sensitive hardcoded keys\""))
	fmt.Println("    " + core.MutedStyle.Sprint("aifiler \"/delete temp log files\""))
	fmt.Println()
}

func (a *App) newClient(providerOverride, modelOverride string) (core.Client, string, string, error) {
	cfg, err := core.LoadOrDefault()
	if err != nil {
		return nil, "", "", fmt.Errorf("%s failed to load config: %w\n  %s Tip: Check permissions or run 'aifiler doctor'", core.ErrorIcon, err, core.InfoIcon)
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

	client := api.NewClient(core.ClientOptions{
		Provider: provider,
		Model:    model,
		Config:   cfg,
	})
	return client, provider, model, nil
}
