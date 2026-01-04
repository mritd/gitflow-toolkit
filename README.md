# GitFlow Toolkit

GitFlow Toolkit is a CLI tool written in Go for standardizing git commit messages following the [Angular commit message specification](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit#heading=h.greljkmo14y0). It provides an interactive TUI for creating commits, branches, and managing git operations.

## Features

- Interactive commit message creation with type, scope, subject, body, and footer
- **AI-powered commit message generation** using LLM (OpenRouter, Groq, OpenAI, or local Ollama)
- Automatic `Signed-off-by` generation
- Git subcommand integration (`git ci`, `git ps`, `git feat`, etc.)
- Lucky commit hash prefix support
- Adaptive terminal UI with light and dark theme support

## Requirements

- Git
- macOS or Linux (Windows is not fully tested)

## Installation

Download the latest binary from the [Release page](https://github.com/mritd/gitflow-toolkit/releases) and run the install command:

```bash
# Download the latest release (replace PLATFORM with: linux-amd64, darwin-arm64, etc.)
curl -fsSL https://github.com/mritd/gitflow-toolkit/releases/latest/download/gitflow-toolkit-PLATFORM -o gitflow-toolkit
chmod +x gitflow-toolkit

# Install (creates symlinks for git subcommands)
sudo ./gitflow-toolkit install
```

Or install via Go:

```bash
go install github.com/mritd/gitflow-toolkit/v3@latest
```

## Usage

After installation, you can use the following git subcommands:

### Commit

```bash
git ci
```

This opens an interactive TUI to create a commit message with:
- Type selection (feat, fix, docs, etc.)
- Scope input
- Subject line
- Optional body (supports external editor with `Ctrl+E`)
- Optional footer

### Push

```bash
git ps
```

Push the current branch to origin with a progress indicator.

### Create Branch

```bash
git feat my-feature    # Creates feat/my-feature
git fix bug-123        # Creates fix/bug-123
git docs readme        # Creates docs/readme
```

## Commands

| Command             | Description                                    |
|---------------------|------------------------------------------------|
| `git ci`            | Interactive commit message creation            |
| `git ps`            | Push current branch to remote                  |
| `git feat NAME`     | Create branch `feat/NAME`                      |
| `git fix NAME`      | Create branch `fix/NAME`                       |
| `git hotfix NAME`   | Create branch `hotfix/NAME`                    |
| `git docs NAME`     | Create branch `docs/NAME`                      |
| `git style NAME`    | Create branch `style/NAME`                     |
| `git refactor NAME` | Create branch `refactor/NAME`                  |
| `git chore NAME`    | Create branch `chore/NAME`                     |
| `git perf NAME`     | Create branch `perf/NAME`                      |
| `git test NAME`     | Create branch `test/NAME`                      |

## Commit Message Format

The tool enforces the Angular commit message format:

```
type(scope): subject

body

footer

Signed-off-by: Name <email>
```

**Supported types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `hotfix`

## Configuration

All settings are configured via `~/.gitconfig` under the `[gitflow]` section.

```ini
[gitflow]
    # LLM API key (required for cloud providers)
    llm-api-key = sk-or-v1-xxxxx
    
    # LLM settings
    llm-host = https://openrouter.ai
    llm-model = mistralai/devstral-2512:free
    llm-temperature = 0.3
    llm-context = 5
    llm-timeout = 120
    llm-retries = 0
    llm-lang = en
    llm-concurrency = 5
    
    # Custom prompts (optional, language-specific)
    llm-file-prompt = "Summarize this diff briefly."
    llm-commit-prompt-en = "Your custom English commit prompt."
    llm-commit-prompt-zh = "Your custom Chinese commit prompt."
    llm-commit-prompt-bilingual = "Your custom bilingual commit prompt."
    
    # Lucky commit prefix (hex characters, max 16)
    lucky-commit = abc
    
    # SSH strict host key checking (default: false)
    ssh-strict-host-key = false
```

### Configuration Reference

| Key | Description | Default |
|-----|-------------|---------|
| `llm-api-key` | API key for cloud LLM providers | - |
| `llm-host` | LLM API endpoint | see below |
| `llm-model` | LLM model name | see below |
| `llm-temperature` | Model temperature | `0.3` |
| `llm-context` | Diff context lines | `5` |
| `llm-timeout` | Request timeout (seconds) | `120` |
| `llm-retries` | Retry count on failure | `0` |
| `llm-lang` | Output language (`en`, `zh`, `bilingual`) | `en` |
| `llm-concurrency` | Parallel file analysis limit | `5` |
| `llm-file-prompt` | Custom file analysis prompt | - |
| `llm-commit-prompt-en` | Custom English commit prompt | - |
| `llm-commit-prompt-zh` | Custom Chinese commit prompt | - |
| `llm-commit-prompt-bilingual` | Custom bilingual commit prompt | - |
| `lucky-commit` | Lucky commit hex prefix (max 16 chars) | - |
| `ssh-strict-host-key` | SSH strict host key checking | `false` |

### Auto Generate (AI)

Generate commit messages automatically using LLM:

1. Run `git ci` and press `Tab` to switch to the `Auto Generate` button (or press `a`)
2. Wait for AI to generate the commit message
3. Review the generated message, then choose:
   - **Commit**: Use the message as-is
   - **Edit**: Open in `$EDITOR` for modifications
   - **Retry**: Regenerate the message

**Provider Selection:**

| Provider | When | Default Host | Default Model |
|----------|------|--------------|---------------|
| OpenRouter | API key is set | `https://openrouter.ai` | `mistralai/devstral-2512:free` |
| Groq | API key set + host contains `groq.com` | - | - |
| OpenAI | API key set + host contains `openai.com` | - | - |
| Ollama | No API key | `http://localhost:11434` | `qwen2.5-coder:7b` |

**Quick Start with OpenRouter (recommended):**
```bash
git config --global gitflow.llm-api-key "sk-or-v1-xxxxx"
git ci  # Press 'a' or Tab to Auto Generate
```

**Quick Start with Local Ollama:**
```bash
ollama pull qwen2.5-coder:7b
git ci
```

**Language Options:**
- `en` - English only (default)
- `zh` - Chinese subject and body (type/scope remain English)
- `bilingual` - Bilingual subject `english (中文)` with Chinese body

### Lucky Commit

Generate commit hashes with a specific prefix using [lucky_commit](https://github.com/not-an-aardvark/lucky-commit):

```bash
# Install lucky_commit first
cargo install lucky_commit

# Set the desired prefix (hex characters, max 16)
git config --global gitflow.lucky-commit abc

# Commit as usual - hash will start with "abc"
git ci
```

- Prefix must be valid hex characters (0-9, a-f)
- Maximum prefix length is 16 characters
- Press Ctrl+C during search to skip and keep original commit

## Uninstall

```bash
sudo gitflow-toolkit uninstall
```

## License

MIT
