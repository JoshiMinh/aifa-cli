package cli

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"aifa/internal/config"
	"aifa/internal/llm"
	"aifa/internal/models"
	"aifa/internal/ops"
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

	switch args[0] {
	case "help", "-h", "--help":
		a.printHelp()
		return 0
	case "rename":
		return a.runRename(ctx, args[1:])
	case "organize":
		return a.runOrganize(ctx, args[1:])
	case "metadata":
		return a.runMetadata(ctx, args[1:])
	case "models":
		return a.runModels(ctx, args[1:])
	case "config":
		return a.runConfig(args[1:])
	default:
		errorStyle.Printf("Unknown command: %s\n\n", args[0])
		a.printHelp()
		return 1
	}
}

func (a *App) printHelp() {
	headerStyle.Println("aifa — AI File Assistant")
	fmt.Println("AI-powered, local-first file management copilot.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aifa <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  rename     Suggest and safely apply semantic file renames")
	fmt.Println("  organize   Group files by semantic categories")
	fmt.Println("  metadata   Suggest metadata updates")
	fmt.Println("  models     View curated model registry and local Ollama models")
	fmt.Println("  config     Initialize local configuration")
	fmt.Println()
	fmt.Println("Safety:")
	fmt.Println("  Operations default to dry-run. Use --apply to execute changes.")
}

func (a *App) runRename(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("rename", flag.ContinueOnError)
	path := fs.String("path", ".", "file or folder to process")
	provider := fs.String("provider", "none", "LLM provider (openai|anthropic|google|ollama|none)")
	model := fs.String("model", "", "model name (optional, uses curated default if empty)")
	apply := fs.Bool("apply", false, "apply changes (default is dry-run)")
	recursive := fs.Bool("recursive", true, "walk subfolders when path is a directory")
	maxFiles := fs.Int("max-files", 100, "maximum files to inspect")
	if err := fs.Parse(args); err != nil {
		errorStyle.Println(err)
		return 2
	}

	cfg, _ := config.LoadOrDefault()
	registry, err := models.LoadRegistry(models.DefaultRegistryPath)
	if err != nil {
		errorStyle.Printf("failed to load model registry: %v\n", err)
		return 1
	}

	selectedModel := strings.TrimSpace(*model)
	if selectedModel == "" {
		selectedModel = registry.DefaultModelForProvider(*provider)
	}

	client := llm.NewClient(llm.ClientOptions{
		Provider: *provider,
		Model:    selectedModel,
		Config:   cfg,
	})

	planner := ops.NewRenamePlanner(client)
	target := filepath.Clean(*path)
	dryRun := !*apply

	result, err := planner.Plan(ctx, ops.RenamePlanInput{
		TargetPath: target,
		Recursive:  *recursive,
		MaxFiles:   *maxFiles,
		DryRun:     dryRun,
	})
	if err != nil {
		errorStyle.Printf("rename failed: %v\n", err)
		return 1
	}

	result.Print()
	if dryRun {
		warnStyle.Println("Dry-run only. Re-run with --apply to execute renames.")
	}
	return 0
}

func (a *App) runOrganize(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("organize", flag.ContinueOnError)
	path := fs.String("path", ".", "folder to organize")
	apply := fs.Bool("apply", false, "apply folder moves (default is dry-run)")
	if err := fs.Parse(args); err != nil {
		errorStyle.Println(err)
		return 2
	}

	planner := ops.NewOrganizer()
	result, err := planner.Plan(ctx, ops.OrganizeInput{TargetPath: *path, DryRun: !*apply})
	if err != nil {
		errorStyle.Printf("organize failed: %v\n", err)
		return 1
	}
	result.Print()
	if !*apply {
		warnStyle.Println("Dry-run only. Re-run with --apply to execute moves.")
	}
	return 0
}

func (a *App) runMetadata(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("metadata", flag.ContinueOnError)
	path := fs.String("path", ".", "target path")
	if err := fs.Parse(args); err != nil {
		errorStyle.Println(err)
		return 2
	}

	suggester := ops.NewMetadataSuggester()
	result, err := suggester.Suggest(ctx, *path)
	if err != nil {
		errorStyle.Printf("metadata failed: %v\n", err)
		return 1
	}
	result.Print()
	return 0
}

func (a *App) runModels(ctx context.Context, args []string) int {
	_ = ctx
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	provider := fs.String("provider", "", "optional provider filter")
	if err := fs.Parse(args); err != nil {
		errorStyle.Println(err)
		return 2
	}

	registry, err := models.LoadRegistry(models.DefaultRegistryPath)
	if err != nil {
		errorStyle.Printf("failed to load model registry: %v\n", err)
		return 1
	}

	registry.Print(*provider)
	ollamaModels, err := llm.DetectOllamaModels(context.Background())
	if err == nil && len(ollamaModels) > 0 {
		fmt.Println()
		headerStyle.Println("Detected local Ollama models")
		for _, model := range ollamaModels {
			fmt.Printf("  - %s\n", model)
		}
	}
	return 0
}

func (a *App) runConfig(args []string) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	initFlag := fs.Bool("init", false, "create a default config file")
	if err := fs.Parse(args); err != nil {
		errorStyle.Println(err)
		return 2
	}

	if !*initFlag {
		fmt.Println("Usage: aifa config --init")
		return 0
	}

	path, err := config.InitDefault()
	if err != nil {
		errorStyle.Printf("failed to init config: %v\n", err)
		return 1
	}
	successStyle.Printf("Config created: %s\n", path)
	return 0
}
