// Package consts defines constants for gitflow-toolkit.
package consts

import "time"

// Commit types following Angular commit message specification.
const (
	Feat     = "feat"
	Fix      = "fix"
	Docs     = "docs"
	Style    = "style"
	Refactor = "refactor"
	Test     = "test"
	Chore    = "chore"
	Perf     = "perf"
	Hotfix   = "hotfix"
)

// Command aliases for git subcommands.
const (
	CmdCommit = "ci"
	CmdPush   = "ps"
)

// CommitType represents a commit type with its name and description.
type CommitType struct {
	Name        string
	Description string
}

// CommitTypes returns all available commit types with descriptions.
var CommitTypes = []CommitType{
	{Feat, "Introducing new features"},
	{Fix, "Bug fix"},
	{Docs, "Writing docs"},
	{Style, "Improving structure/format of the code"},
	{Refactor, "Refactoring code"},
	{Test, "When adding missing tests"},
	{Chore, "Changing CI/CD"},
	{Perf, "Improving performance"},
	{Hotfix, "Bug fix urgently"},
}

// Lucky commit constants.
const (
	// LuckyCommitBinary is the name of the lucky_commit executable.
	LuckyCommitBinary = "lucky_commit"

	// LuckyCommitMaxLen is the maximum length of lucky commit prefix.
	LuckyCommitMaxLen = 12

	// LuckyCommitURL is the download URL for lucky_commit.
	LuckyCommitURL = "https://github.com/not-an-aardvark/lucky-commit"
)

// LLM defaults.
const (
	LLMDefaultOllamaHost     = "http://localhost:11434"
	LLMDefaultOpenRouterHost = "https://openrouter.ai"
	LLMDefaultDiffContext    = 5
	LLMDefaultRequestTimeout = 2 * time.Minute
	LLMDefaultRetries        = 0
	LLMDefaultLang           = "en"
	LLMDefaultTemperature    = 0.3
	LLMDefaultConcurrency    = 3
)

// LLM language options.
const (
	LLMLangEN        = "en"
	LLMLangZH        = "zh"
	LLMLangBilingual = "bilingual"
)

// LLM default models.
const (
	LLMDefaultOllamaModel     = "qwen2.5-coder:7b"
	LLMDefaultOpenRouterModel = "mistralai/devstral-2512:free"
)

// LLM default prompts (can be overridden via gitconfig).
const (
	// LLMDefaultFilePrompt is the system prompt for analyzing individual file diffs.
	LLMDefaultFilePrompt = "You are a git diff analyzer. Output only a brief summary, no formatting."

	// LLMCommitPromptEN is the system prompt for generating commit messages in English.
	LLMCommitPromptEN = `You are a git commit message generator. Generate EXACTLY ONE commit message following the Angular commit convention.

FORMAT (strict):
<type>(<scope>): <subject>

<body>

RULES:
1. type: REQUIRED, one of: feat, fix, docs, style, refactor, test, chore, perf, hotfix
2. scope: REQUIRED, a short word describing the affected area (e.g., api, ui, config, auth, db, cli)
3. subject: REQUIRED, imperative mood, lowercase, no period, max 50 chars
4. body: REQUIRED, 3-5 bullet points starting with "- ", each point starts with a verb

OUTPUT ONLY THE COMMIT MESSAGE. No explanation, no markdown, no code blocks.`

	// LLMCommitPromptZH is the system prompt for generating commit messages in Chinese.
	LLMCommitPromptZH = `你是一个 git commit 消息生成器.请严格按照 Angular commit 规范生成一条 commit 消息.

格式要求（严格遵守）:
<type>(<scope>): <中文描述>

<正文>

规则:
1. type: 必填, 只能是: feat, fix, docs, style, refactor, test, chore, perf, hotfix
2. scope: 必填, 描述影响范围的英文单词（如 api, ui, config, auth, db, cli）
3. subject: 必填, 使用中文描述, 不加句号, 最多50字
4. body: 必填, 3-5个要点, 每行以"- "开头, 使用中文描述

只输出 commit 消息本身, 不要任何解释, markdown 或代码块.`

	// LLMCommitPromptBilingual is the system prompt for generating bilingual commit messages.
	LLMCommitPromptBilingual = `You are a git commit message generator. Generate EXACTLY ONE commit message following the Angular commit convention with bilingual subject.

FORMAT (strict):
<type>(<scope>): <english subject> (<中文描述>)

<body in Chinese>

RULES:
1. type: REQUIRED, one of: feat, fix, docs, style, refactor, test, chore, perf, hotfix
2. scope: REQUIRED, a short word describing the affected area (e.g., api, ui, config, auth, db, cli)
3. subject: REQUIRED, format "english description (中文描述)", lowercase English, no period
4. body: REQUIRED, 3-5 bullet points starting with "- ", written in Chinese

OUTPUT ONLY THE COMMIT MESSAGE. No explanation, no markdown, no code blocks.`
)

// Binary and path constants.
const (
	// BinaryName is the name of the main binary.
	BinaryName = "gitflow-toolkit"

	// GitCommandPrefix is the prefix for git subcommands.
	GitCommandPrefix = "git-"

	// DefaultInstallDir is the default installation directory.
	DefaultInstallDir = "/usr/local/bin"

	// TempFilePrefix is the prefix for temporary files.
	TempFilePrefix = "gitflow"
)

// SymlinkCommands returns all symlink command names (without git- prefix).
func SymlinkCommands() []string {
	return []string{
		CmdCommit,
		CmdPush,
		Feat,
		Fix,
		Docs,
		Style,
		Refactor,
		Test,
		Chore,
		Perf,
		Hotfix,
	}
}
