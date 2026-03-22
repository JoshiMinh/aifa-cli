# aifiler — AI File Assistant 🚀

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-6E40C9)](#quick-start)
[![CLI](https://img.shields.io/badge/Type-Command%20Line-1F883D)](#commands)

`aifiler` is a local-first CLI assistant for AI-powered file and folder operations.
It can generate create/rename plans from prompts, execute those plans in your current directory, and switch between providers/models quickly.

---

## ✨ Features

- 🧠 Prompt-based file/folder creation (`create`)
- 🏷️ Prompt-based file/folder renaming (`rename`)
- 💬 Free-form dynamic prompts (`aifiler "..."`)
- 🗂️ Automatic workspace structure context (no manual folder description needed)
- ✅ Approval-gated execution for file/folder updates and shell commands
- 🔌 Multi-provider support (`ollama`, `vercel`, fallback `none`)
- 📚 Curated model registry + auto-detected model lists
- ⚙️ Simple provider key and default model management

---

## 🛠️ Quick Start

### Windows (PowerShell)

We provide handy PowerShell wrapper scripts to compile and run the CLI smoothly:

```powershell
# Install dependencies
go mod tidy

# Build the executable (creates aifiler.exe)
.\build.ps1

# Run the executable (or builds it if missing)
.\run.ps1 help
```

### macOS/Linux

```bash
# Install dependencies
go mod tidy

# Build the executable
go build -o aifiler ./cmd/aifiler

# Run the CLI
./aifiler help
```

---

## 📦 Installation / Build Notes

- Requires Go `1.22+`
- Registry lookup order: `%AIFILER_MODEL_REGISTRY%` → current folder → executable folder
- Default config file location:
  - Windows: `%AppData%/aifiler/config.yaml`
  - macOS/Linux: `$XDG_CONFIG_HOME/aifiler/config.yaml` (or `~/.config/aifiler/config.yaml`)

---

## 🧭 Commands

| Command                                | Purpose                                                     |
| -------------------------------------- | ----------------------------------------------------------- |
| `aifiler` / `aifiler help`         | Show help                                                   |
| `aifiler list`                       | List providers, models, API key status, and detected models |
| `aifiler set "provider" "api key"`   | Save provider API key                                       |
| `aifiler default "model"`            | Set default model                                           |
| `aifiler reset "provider" "api key"` | Remove/reset provider API key                               |
| `aifiler doctor`                     | Show runtime diagnostics (registry resolution, paths)       |
| `aifiler create "<prompt>"`          | Create files/folders from AI plan                           |
| `aifiler rename "<prompt>"`          | Rename files/folders from AI plan                           |
| `aifiler "<prompt>"`                 | Run dynamic prompt; can propose/execute actions with approval |

Notes:
- Prompts automatically include a snapshot of the current working folder structure.
- Any mutating action (create/update/rename/run command) is shown first and only executed after user approval.
- At the proposal prompt, you can type `y`/`n` or directly type the next prompt to refine the plan in-place.

---

## 🔌 Provider Manuals

### 1) Vercel AI Gateway (OpenAI-compatible)

Use provider name: `vercel`

```powershell
aifiler set "vercel" "<your-ai-gateway-api-key>"
aifiler default "openai/gpt-4o-mini"
aifiler "Summarize this repository structure"
```

Also supported via environment variables:

```powershell
$env:AI_GATEWAY_API_KEY = "<your-key>"
$env:AI_GATEWAY_BASE_URL = "https://ai-gateway.vercel.sh/v1" # optional override
```

Examples of gateway model IDs:

- `openai/gpt-4o-mini`
- `anthropic/claude-sonnet-4.5`
- `google/gemini-2.5-flash`

`aifiler list` now attempts to detect available Vercel models via gateway `/models` when credentials are available.

### 2) Ollama (local)

Use provider name: `ollama`

```powershell
aifiler set "ollama" "local"
aifiler default "llama3.2"
aifiler "Create a folder structure for docs"
```

`aifiler list` auto-detects local Ollama models from `http://127.0.0.1:11434`.

---

## 📖 Usage Workflows

### Create files and folders

**Windows:**

```powershell
.\run.ps1 create "create src and docs folders with starter files"
```

**macOS/Linux:**

```bash
./aifiler create "create src and docs folders with starter files"
```

### Rename files and folders

**Windows:**

```powershell
.\run.ps1 rename "rename all markdown files to kebab-case"
```

**macOS/Linux:**

```bash
./aifiler rename "rename all markdown files to kebab-case"
```

### Quick one-off prompt

**Windows:**

```powershell
.\run.ps1 "propose a clean monorepo structure for a Go CLI"
```

**macOS/Linux:**

```bash
./aifiler "propose a clean monorepo structure for a Go CLI"
```

---

## 🧪 Troubleshooting

- **"missing API key for provider 'vercel'"**
  - Run `aifiler set "vercel" "<api-key>"`, or set `AI_GATEWAY_API_KEY`
- **No detected Vercel models**
  - Check key validity, network access, and gateway endpoint
- **No detected Ollama models**
  - Ensure Ollama is running locally and model(s) are installed
- **Registry load error**
  - Set `%AIFILER_MODEL_REGISTRY%` to a valid `registry.yaml` path, or keep `assets/models/registry.yaml` beside the executable
  - Run `aifiler doctor` to see which path is selected at runtime

---

## 🚢 Release & Publish Manual (Scoop / Chocolatey)

### 1) One-time setup

```powershell
go install github.com/goreleaser/goreleaser/v2@latest
$env:GITHUB_TOKEN = "<your-github-token>"
$env:SCOOP_GITHUB_TOKEN = "<token-with-access-to-scoop-bucket>"
```

### 2) Verify release config

- Check `.goreleaser.yaml`
- Confirm `project_name`, scoop repo owner/name, and Chocolatey URL template
- Add `LICENSE` if your manifest requires it

### 3) Tag + release

```powershell
git tag v0.1.0
git push origin v0.1.0
goreleaser release --clean
```

### 4) Install from Scoop

```powershell
scoop bucket add JoshiMinh https://github.com/JoshiMinh/scoop-bucket
scoop install aifiler
```

### 5) Install from Chocolatey

```powershell
choco install aifiler
```

Local package test:

```powershell
choco install aifiler --source .
```

---

## 🗺️ Roadmap Ideas

- Provider-specific model filters in `list`
- Optional dry-run mode for `create`/`rename`
- JSON schema validation for AI-generated plans

---

Built with ❤️ for fast terminal workflows.
