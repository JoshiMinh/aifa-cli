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
type App struct {
	maxDepth int
	showAll  bool
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		maxDepth: 1, // Default: root only
		showAll:  false,
	}
}

// Run executes the CLI application with the given arguments.
// It returns the exit code.
func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.printHelp()
		return 0
	}

	// Pre-process global flags: -r, -ra, -all
	var remainingArgs []string
	for _, arg := range args {
		switch arg {
		case "-r":
			a.maxDepth = 2
		case "-ra":
			a.maxDepth = 0 // recursive
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
	case "history":
		return a.runHistory()
	case "undo":
		return a.runUndo()
	case "set":
		return a.runSet(remainingArgs[1:])
	case "default":
		return a.runDefault(remainingArgs[1:])
	case "reset":
		return a.runReset(remainingArgs[1:])
	default:
		// All other inputs (or starting with /) go to dynamic prompt
		return a.runDynamicPrompt(ctx, strings.Join(remainingArgs, " "))
	}
}

func (a *App) printHelp() {
	headerStyle.Println("\n  " + sparkleIcon + " aifiler — AI-Powered Filesystem Assistant")
	fmt.Println("  --------------------------------------------------")
	fmt.Println("  A local-first assistant that translates natural language into safely")
	fmt.Println("  orchestrated filesystem operations.")
	fmt.Println()

	headerStyle.Println("  MAIN COMMAND")
	fmt.Printf("    %-25s %s\n", pathStyle.Sprint("aifiler \"<prompt>\""), "Execute an AI-powered plan based on your request")
	fmt.Printf("    %-25s %s\n", pathStyle.Sprint("aifiler \"/<intent> ...\""), "Force a specific operation (e.g., /create, /rename, /delete)")
	fmt.Println()

	headerStyle.Println("  OPTIONS")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("-r"), "Scan root and immediate subfolders (one-level)")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("-ra"), "Fully recursive scan (all levels)")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("-all"), "Include all file entries in AI context (no truncation)")
	fmt.Println()

	headerStyle.Println("  UTILITIES")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("history"), "View recent AI operations")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("undo"), "Revert the last applied AI plan")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("list"), "Model selection (Vercel recommended)")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("doctor"), "Show runtime diagnostics")
	fmt.Println()

	headerStyle.Println("  CONFIGURATION")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("set \"provider\""), "Save API key securely")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("default \"model\""), "Set default model")
	fmt.Printf("    %-25s %s\n", mutedStyle.Sprint("reset \"provider\""), "Remove saved key")
	fmt.Println()

	headerStyle.Println("  EXAMPLES")
	fmt.Println("    " + mutedStyle.Sprint("aifiler \"create a clean src layout with README\""))
	fmt.Println("    " + mutedStyle.Sprint("aifiler -r \"rename all .js files to .ts\""))
	fmt.Println("    " + mutedStyle.Sprint("aifiler -ra -all \"search for sensitive hardcoded keys\""))
	fmt.Println("    " + mutedStyle.Sprint("aifiler \"/delete temp log files\""))
	fmt.Println()
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
