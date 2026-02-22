# aifa — AI File Assistant

aifa is a local-first CLI copilot for semantic file management.

## Core capabilities (MVP scaffold)

- Intelligent rename planning (`aifa rename`)
- Organization planning/grouping (`aifa organize`)
- Metadata suggestions (`aifa metadata`)
- Safe by default: dry-run mode unless `--apply` is set
- Curated static model registry in YAML (`assets/models/registry.yaml`)
- Local Ollama model auto-detection (`aifa models`)

## Why this structure

- Provider-agnostic LLM interface in [internal/llm](internal/llm)
- Composable operations in [internal/ops](internal/ops)
- Simple CLI composition in [internal/cli](internal/cli)
- Local config in user config directory (`aifa/config.yaml`)

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
