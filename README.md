# aifiler — AI File Assistant 🚀

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%7CmacOS%7CLinux-6E40C9)](#quick-start)

`aifiler` is a local-first CLI assistant for AI-powered file/folder operations. It generates plans from prompts, executes them with approval, and maintains a clean, searchable workspace.

---

## ✨ Features

- 🧠 **Smart Plans**: Generate file structural changes from natural language.
- 🗂️ **Context Aware**: Automatically recognizes current directory structure.
- ✅ **Safety First**: All destructive actions require explicit user approval.
- 🔌 **Unified Interface**: Supports major text providers (OpenAI, Google, Anthropic, etc.).
- 📊 **Beautiful CLI**: Rich output with icons, colors, and tabular file listings.

---

## 🛠️ Quick Start (Windows)

```batch
:: Build the executable (includes go mod tidy)
build.bat

:: Run the CLI
run.bat help
```

### macOS / Linux

```bash
go mod tidy
go build -o aifiler ./cmd/aifiler
./aifiler help
```

---

## 🧭 Commands

| Command | Description |
| :--- | :--- |
| `aifiler help` | Show usage information |
| `aifiler ls [-r, -ra]` | Tabular file list (root, level-1, or recursive) |
| `aifiler list` | Show approved text providers and models |
| `aifiler set "p"` | Securely prompt and save provider API key |
| `aifiler default "m"` | Set default model |
| `aifiler reset "p"` | Remove provider API key |
| `aifiler create "..."` | Create files/folders from AI plan |
| `aifiler rename "..."` | Rename files from AI suggestions |
| `aifiler "..."` | Run dynamic AI prompt |

---

## 🔌 Setup Examples

### OpenAI / Anthropic

```batch
run.bat set "openai"
run.bat default "gpt-4o-mini"
run.bat "create a clean src layout for a python app"
```

---

## 🧪 Troubleshooting

- **Missing Key**: Run `aifiler set <provider> <key>`.
- **Registry Error**: Run `aifiler doctor` to check path resolution.
- **Recursion**: Use `ls -r` for subfolders or `ls -ra` for deep scan.

---

Built with ❤️ for terminal productivity.
