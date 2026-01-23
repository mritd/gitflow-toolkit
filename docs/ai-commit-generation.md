# AI Commit Message Generation

This document describes the AI-powered commit message generation flow, including file analysis, prioritization, and debug logging.

## Overview

AI commit generation uses a two-phase approach:

1. **Phase 1 (File Analysis)**: Analyze each file's changes with global context
2. **Phase 2 (Commit Generation)**: Generate final commit message from sorted summaries

## Generation Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AI Commit Generation Flow                         │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐
│  git ci → 'a'   │  User triggers AI generation
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 0: Preprocessing                                                      │
│                                                                             │
│  1. Get staged diff (git diff --cached)                                     │
│  2. Split diff by file                                                      │
│  3. Filter lock files (if >= 5 files)                                       │
│  4. Truncate large diffs by priority                                        │
│  5. Detect CORE files (top 3 code files by priority + lines)                │
│  6. Sort files: [CORE] code > code > test > config > doc > other            │
└────────┬────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 1: File Analysis (Parallel)                                           │
│                                                                             │
│  For each file (max concurrency from config):                               │
│                                                                             │
│    ┌─────────────────────────────────────────────────────────────────────┐  │
│    │ Prompt Structure:                                                   │  │
│    │                                                                     │  │
│    │   This commit changes N files:                                      │  │
│    │   [CORE] internal/ui/ai.go (+100/-50)                               │  │
│    │         internal/git/filter.go (+30/-10)                            │  │
│    │         README.md (+5/-2)                                           │  │
│    │                                                                     │  │
│    │   Now analyze [CORE]: internal/ui/ai.go                             │  │
│    │   <diff content>                                                    │  │
│    │                                                                     │  │
│    │   Describe what changed and why (max 50 words)...                   │  │
│    └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  Output: Summary for each file (max 50 words)                               │
└────────┬────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 2: Commit Message Generation                                          │
│                                                                             │
│    ┌─────────────────────────────────────────────────────────────────────┐  │
│    │ Prompt Structure:                                                   │  │
│    │                                                                     │  │
│    │   Few-shot examples (language-specific)...                          │  │
│    │                                                                     │  │
│    │   Input:                                                            │  │
│    │   [CORE] internal/ui/ai.go: Add two-phase AI generation flow        │  │
│    │         internal/git/filter.go: Add file sorting functions          │  │
│    │         README.md: Update AI generation documentation               │  │
│    │                                                                     │  │
│    │   Output (MUST include body with "- " lines):                       │  │
│    └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  Output: Complete commit message in Angular format                          │
└────────┬────────────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  Preview & Edit │  User reviews, edits, or retries
└─────────────────┘
```

## File Prioritization

### CORE File Detection

CORE files are the most important files in a commit. Detection considers:

1. **File Type Priority** (higher priority wins):
   - Code files (`.go`, `.py`, `.js`, `.ts`, etc.) - priority 0
   - Config files (`go.mod`, `package.json`, `.yaml`, etc.) - priority 1
   - Doc files (`.md`, `.txt`, `.rst`, etc.) - priority 2
   - Other files (`.gitignore`, etc.) - priority 3
   - Test files - excluded from CORE detection

2. **Lines Changed**: Within same priority, more lines changed = more important

3. **Selection**: Top 3 files by (priority, lines) become CORE files

Example:
```
Files in commit:
  internal/ui/ai.go      (+100/-50)  → code, 150 lines  → CORE
  README.md              (+500/-200) → doc, 700 lines   → not CORE (doc < code)
  internal/git/filter.go (+80/-30)   → code, 110 lines  → CORE
  main_test.go           (+200/-100) → test, excluded
  go.mod                 (+5/-2)     → config, 7 lines  → CORE (fills remaining slot)
```

### File Sorting Order

Files are sorted for LLM attention (important files first):

1. `[CORE]` code files (by lines changed, descending)
2. Regular code files
3. Test files
4. Config files
5. Documentation files
6. Other files

### Lock File Filtering

When commit has >= 5 files, lock files are automatically filtered:
- `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`
- `go.sum`, `Cargo.lock`, `Gemfile.lock`
- `poetry.lock`, `Pipfile.lock`, `uv.lock`, `pdm.lock`
- `composer.lock`, `pubspec.lock`, `packages.lock.json`

### Diff Truncation

Large diffs are truncated by priority to fit within `llm-max-diff-lines` (default: 500):

1. High priority files (source code) - preserved first
2. Medium priority files (config/scripts) - preserved second
3. Low priority files (tests/docs) - truncated first

Truncated files show: `(file changes: +N/-M lines, content omitted)`

## Debug Logging

### Enable Debug Mode

```bash
git config --global gitflow.llm-api-debug true
```

### Debug Log Location

The debug log is written to:
- **macOS/Linux**: `$TMPDIR/gitflow-llm-debug.log` (typically `/var/folders/.../gitflow-llm-debug.log`)
- Use `echo $TMPDIR` to find the exact path

After AI generation, the log path is displayed in the preview UI.

### Debug Log Contents

The debug log records each phase of AI generation:

```
================================================================================
Phase 0: Original File List
================================================================================
Total files: 5

 1. internal/ui/commit/ai.go (+100/-50)
 2. internal/git/filter.go (+30/-10)
 3. README.md (+5/-2)
 4. go.mod (+2/-1)
 5. .gitignore (+1/-0)

================================================================================
Phase 0: Core Files Detection
================================================================================
Core files (top 3 code files by lines changed):
  - internal/ui/commit/ai.go (+100/-50)
  - internal/git/filter.go (+30/-10)

================================================================================
Phase 0: Sorted File List
================================================================================
Sorted by priority: [CORE] code > code > test > config > doc > other

 1. [CORE] internal/ui/commit/ai.go (+100/-50)
 2. [CORE] internal/git/filter.go (+30/-10)
 3.        go.mod (+2/-1)
 4.        README.md (+5/-2)
 5.        .gitignore (+1/-0)

[2024-01-23 15:30:01] Phase 1: [1/5] internal/ui/commit/ai.go => Add two-phase AI generation with global context
[2024-01-23 15:30:02] Phase 1: [2/5] internal/git/filter.go => Add file sorting and CORE detection functions
[2024-01-23 15:30:03] Phase 1: [3/5] go.mod => Update module dependencies
[2024-01-23 15:30:04] Phase 1: [4/5] README.md => Document AI generation feature
[2024-01-23 15:30:05] Phase 1: [5/5] .gitignore => Add debug log to gitignore

================================================================================
Phase 1 Complete: File Summaries
================================================================================
All file summaries:

[CORE] internal/ui/commit/ai.go:
  Add two-phase AI generation with global context

[CORE] internal/git/filter.go:
  Add file sorting and CORE detection functions

       go.mod:
  Update module dependencies

       README.md:
  Document AI generation feature

       .gitignore:
  Add debug log to gitignore

================================================================================
Phase 2 Complete: Generated Commit Message
================================================================================
feat(ai): add two-phase AI commit generation

- Add global context for file analysis
- Add file sorting by priority
- Add CORE file detection
```

### LLM API Request/Response Logging

When debug mode is enabled, each LLM API request and response is also logged:

```
================================================================================
LLM Request
================================================================================
Provider: openrouter
Model: mistralai/devstral-2512:free
Endpoint: https://openrouter.ai/api/v1/chat/completions

System Prompt:
You are analyzing a code change...

User Prompt:
This commit changes 5 files:
[CORE] internal/ui/ai.go (+100/-50)
...

================================================================================
LLM Response
================================================================================
Add two-phase AI generation flow with global context awareness
```

### Disable Debug Mode

```bash
git config --global --unset gitflow.llm-api-debug
```

## Configuration Reference

| Key | Description | Default |
|-----|-------------|---------|
| `llm-api-debug` | Enable debug logging | `false` |
| `llm-max-diff-lines` | Max diff lines to analyze | `500` |
| `llm-max-concurrency` | Parallel file analysis | `3` |
| `llm-file-analysis-prompt` | Custom file analysis prompt | (built-in) |
| `llm-commit-prompt-en` | Custom English commit prompt | (built-in) |
| `llm-commit-prompt-zh` | Custom Chinese commit prompt | (built-in) |
| `llm-commit-prompt-bilingual` | Custom bilingual commit prompt | (built-in) |

## Troubleshooting

### AI generates incorrect commit messages

1. Enable debug mode to inspect the prompts and responses
2. Check if important files are marked as `[CORE]`
3. Verify file sorting puts code files first
4. Check if diffs are being truncated (look for "content omitted")

### File not appearing in analysis

1. Check if file is a lock file (filtered when >= 5 files)
2. Check if diff exceeds `llm-max-diff-lines` limit
3. Verify file is staged (`git status`)

### Slow generation

1. Reduce `llm-max-concurrency` if rate limited
2. Reduce `llm-max-diff-lines` for faster processing
3. Check network connectivity to LLM provider
