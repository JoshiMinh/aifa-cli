# aifiler — AI File Assistant

aifiler is a command-line assistant for AI-assisted file and folder operations.
It supports prompt-driven create/rename workflows, dynamic prompting, and provider/model configuration.

## Quick start

```bash
go mod tidy
# Windows (PowerShell)
./build.ps1
./run.ps1 --help

# macOS/Linux
go build -o aifiler ./cmd/aifiler
./aifiler --help
```

## Release & publish (Scoop/Chocolatey)

### 1) One-time setup

- Install GoReleaser: `go install github.com/goreleaser/goreleaser/v2@latest`
- Create a GitHub Personal Access Token with repo write access and set it:

```powershell
$env:GITHUB_TOKEN = "<your-github-token>"
```

- Create a Scoop bucket repo (if you don't already have one): `JoshiMinh/scoop-bucket`
- Create another token for bucket updates and set it:

```powershell
$env:SCOOP_GITHUB_TOKEN = "<token-with-access-to-scoop-bucket>"
```

### 2) Prepare release metadata

- GoReleaser config is at `.goreleaser.yaml`
- Verify `project_name`, `scoops.repository.owner/name`, and `chocolateys.url_template`
- If you keep `license: MIT` in scoop manifest, add a `LICENSE` file in this repo

### 3) Tag and publish with GoReleaser

```powershell
git tag v0.1.0
git push origin v0.1.0
goreleaser release --clean
```

This creates GitHub release artifacts and updates your Scoop bucket manifest.

### 4) Install using Scoop

```powershell
scoop bucket add JoshiMinh https://github.com/JoshiMinh/scoop-bucket
scoop install aifiler
```

### 5) Install using Chocolatey

If you publish to Chocolatey Community Repository and package is approved:

```powershell
choco install aifiler
```

For local/manual package testing from generated `.nupkg`:

```powershell
choco install aifiler --source .
```

### 6) Verify install

```powershell
aifiler --help
```

## Commands

- `aifiler` / `aifiler help`: list commands
- `aifiler create "<prompt>"`: create files/folders from AI suggestions
- `aifiler rename "<prompt>"`: rename files/folders from AI suggestions
- `aifiler "<prompt>"`: dynamic AI response without explicit command
- `aifiler list`: list available providers/models and configured API key status
- `aifiler set "provider" "api key"`: save API key for provider
- `aifiler default "model"`: set default model
- `aifiler reset "provider" "api key"`: reset/remove provider key

## Example commands

```bash
aifiler
aifiler help
aifiler list
aifiler "organize this repository structure"
aifiler set "ollama" "your-api-key"
aifiler default "llama3.2"
aifiler reset "ollama" "your-api-key"
aifiler create "create src and docs folders with starter files"
aifiler rename "rename all docs files to kebab-case"
```
