# aifa — AI File Assistant

aifa is a command-line assistant that helps you clean up messy files and folders.
It looks at what files are (not just what they are named), then suggests better names, groups related items, and helps you keep things organized safely.

## What aifa helps with

- Renaming files with clearer, more consistent names
- Grouping files into useful folders (documents, images, code, and more)
- Suggesting metadata you can review before making changes
- Previewing changes first (dry-run) so nothing happens unexpectedly
- Working locally on your machine, with optional cloud or local AI models

## Why people use it

- Save time when folders become hard to manage
- Keep naming and organization consistent across projects
- Stay in control with safe previews before applying changes
- Use your own AI provider and API keys

## Quick start

```bash
go mod tidy
go build -o aifa ./cmd/aifa
./aifa --help
```

## Packaging

The project includes a starter `.goreleaser.yaml` with Scoop and Chocolatey entries.
Replace placeholder GitHub org/repo values before publishing.

## Safety model

- Destructive operations require explicit `--apply`
- Dry-run previews show planned actions
- Existing targets are never overwritten

## LLM integration note

This scaffold keeps LLM logic provider-agnostic and wires Ollama for local use.
`github.com/hoangvvo/llm-sdk` currently publishes no Go source package, so direct Go import is not available yet.
The integration seam is isolated in `internal/llm` so switching to official `llm-sdk` Go bindings is a focused follow-up when they are published.

## Example commands

```bash
aifa config --init
aifa models
aifa rename --path .
aifa rename --path . --provider ollama --model llama3.2 --apply
aifa organize --path ~/Downloads
aifa metadata --path ~/Documents
```
