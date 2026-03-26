# 🚀 aifiler — AI-powered Filesystem Assistant

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%7CmacOS%7CLinux-6E40C9)](#quick-start)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`aifiler` is a local-first CLI assistant designed for seamless, AI-driven filesystem management. Instead of manual sorting and naming, just describe your intent and let `aifiler` handle the plan, approval, and execution.

---

## ✨ Key Features

- 🧠 **Dynamic Planning**: Translates natural language into structured filesystem operations.
- 🗂️ **Context Awareness**: Intelligently scans your workspace to provide relevant suggestions.
- ✅ **Safety First**: Every action is staged for your approval before execution.
- 🔌 **Provider Agnostic**: Supports OpenAI, Anthropic, Google Gemini, and Ollama via a unified interface.
- 🎨 **Modern CLI**: Rich, colorful output with interactive elements for a premium terminal experience.
- ⏪ **One-Click Undo**: Accidentally applied a plan? Revert changes instantly with the `undo` command.

---

## Quick Start

### Windows

```batch
:: Build the executable
build.bat

:: Get started with help
run.bat help
```

### macOS / Linux

```bash
go mod tidy
go build -o aifiler ./cmd/aifiler
./aifiler help
```

---

## 🧭 Usage Reference

| Command | Action |
| :--- | :--- |
| `aifiler "<prompt>"` | Generate and execute an AI plan from natural language |
| `aifiler "/<intent> ..."` | Force a specific intent (e.g., `/create`, `/rename`, `/delete`) |
| `aifiler -r` | Extend context scan to immediate subfolders (one-level) |
| `aifiler -ra` | Fully recursive scan for deep project context |
| `aifiler history` | View the log of recent operations |
| `aifiler undo` | Revert the last applied plan |
| `aifiler list` | Show available LLM providers and models |

---

## 🔌 Configuration

Get started by setting up your preferred AI provider:

```batch
run.bat set "openai"
run.bat default "gpt-4o-mini"
run.bat "create a clean src layout for a python app"
```

---

## 🏗️ Architecture

`aifiler` is built with modularity and safety in mind:

- **CLI Layer**: Handles input/output using a rich styling system.
- **Plan Engine**: Stages operations and manages user approval.
- **LLM Core**: Abstracted client logic supporting multiple providers.
- **State Manager**: Maintains history for audit logs and undo functionality.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

Built with ❤️ for terminal productivity.
