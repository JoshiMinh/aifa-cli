<div align="center">

# aifiler

AI-powered local Filesystem Assistant. Instead of manual sorting and naming, just describe your intent and let `aifiler` handle the plan, approval, and execution.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%7CmacOS%7CLinux-6E40C9)](#quick-start)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

## Preview

![aifiler preview](preview.png)

---

## ✨ Key Features

- 🧠 **Dynamic Planning**: Translates natural language into structured filesystem operations.
- 🗂️ **Context Awareness**: Intelligently scans your workspace to provide relevant suggestions.
- ✅ **Safety First**: Every action is staged for your approval before execution.
- 🔌 **Provider Agnostic**: Supports OpenAI, Anthropic, Gemini, Ollama, and Vercel AI Gateway via a unified interface.
- 🎨 **Modern CLI**: Clean, professional bold output with interactive arrow-key menus.
- ⏪ **One-Click Undo**: Accidentally applied a plan? Revert changes instantly with the `undo` command.

---

## Quick Start

### Windows

```batch
:: Start the interactive menu (Build/Run)
run.bat

:: Or run directly with arguments
run.bat "organize these files"
```

### macOS / Linux

```bash
go mod tidy
go build -o aifiler ./cmd/aifiler
./aifiler
```

---

## 🧭 Usage Reference

| Command | Action |
| :--- | :--- |
| `aifiler "<prompt>"` | Generate and execute an AI plan from natural language |
| `aifiler "/<intent> ..."` | Force a specific intent (e.g., `/create`, `/rename`, `/delete`) |
| `aifiler -r` | Extend context scan to immediate subfolders (one-level) |
| `aifiler -ra` | Fully recursive scan for deep project context |
| `aifiler provider` | Configure provider: switch active, set API keys, clear keys |
| `aifiler list` | List and set the default model for the active provider |
| `aifiler history` | View the log of recent operations |
| `aifiler undo` | Revert the last applied plan |

---

## 🔌 Configuration

Get started by setting up your preferred AI provider:

```batch
:: Configure your provider (Interactive Menu)
run.bat provider

:: List available models for your active provider
run.bat list

:: Run a prompt
run.bat "organize these messy log files into a logs/ folder"
```

Supported providers (in order of preference):

| Provider | Key | Notes |
| :--- | :--- | :--- |
| OpenAI | `openai` | GPT-4o, o3, and more |
| Anthropic | `anthropic` | Claude 3.x family |
| Gemini | `gemini` | Google Gemini models |
| Ollama | `ollama` | Local models, no API key needed |
| Vercel AI Gateway | `vercel` | Routes to multiple providers |

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
