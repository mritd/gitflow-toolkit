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
	LLMDefaultDiffContext    = 5
	LLMDefaultRequestTimeout = 2 * time.Minute
	LLMDefaultRetries        = 0
	LLMDefaultLang           = "en"
	LLMDefaultTemperature    = 0.3
	LLMDefaultConcurrency    = 3
	LLMDefaultMaxDiffLines   = 500
)

// LLM provider hosts (OpenAI-compatible APIs only).
const (
	LLMHostOllama     = "http://localhost:11434"
	LLMHostOpenRouter = "https://openrouter.ai"
	LLMHostGroq       = "https://api.groq.com"
	LLMHostOpenAI     = "https://api.openai.com"
	LLMHostDeepSeek   = "https://api.deepseek.com"
	LLMHostMistral    = "https://api.mistral.ai"
)

// LLM API paths for chat completions.
const (
	LLMPathOllama     = "/api/generate"
	LLMPathOpenAI     = "/v1/chat/completions"        // OpenAI, DeepSeek, Mistral, most compatible APIs
	LLMPathOpenRouter = "/api/v1/chat/completions"    // OpenRouter
	LLMPathGroq       = "/openai/v1/chat/completions" // Groq
)

// LLM language options.
const (
	LLMLangEN        = "en"
	LLMLangZH        = "zh"
	LLMLangBilingual = "bilingual"
)

// LLM default models.
const (
	LLMModelOllama     = "qwen2.5-coder:7b"
	LLMModelOpenRouter = "mistralai/devstral-2512:free"
)

// LLM default prompts (can be overridden via gitconfig).
const (
	// LLMDefaultFilePrompt is the system prompt for analyzing individual file diffs.
	LLMDefaultFilePrompt = `Summarize the code change in MAX 10 words.

Rules:
- Start with a verb (Add, Update, Fix, Remove, Rename, Refactor, Migrate)
- Describe the ACTION and PURPOSE, not literal strings
- For import/module path changes: "Migrate module path to new repository" (not actual URLs)
- For config changes: "Add config key" or "Rename config key"
- NEVER include actual URLs, domains, or paths in output

Good: "Add debug logging support"
Good: "Migrate module path to new repository"
Good: "Rename config key for clarity"
Bad: "Changed from gitlab.com to git.example.com" (contains literal URL)
Bad: "Update import from old/path to new/path" (contains literal path)

Output ONLY the summary.`

	// LLMCommitPromptEN is the system prompt for generating commit messages in English.
	LLMCommitPromptEN = `Generate ONE commit message in Angular format.

FORMAT (MUST follow exactly):
<type>(<scope>): <subject>

- <body line 1>
- <body line 2 if needed>

RULES:
- type: feat|fix|docs|refactor|test|chore|perf
- scope: the main component being changed (llm, config, ui, git, api)
- subject: what this commit does, max 50 chars, no period
- body: 1-3 lines MAX, each starting with "- "

CRITICAL:
- BODY IS MANDATORY - at least one line starting with "- "
- SUMMARIZE, don't enumerate - if 20 files have same change, write ONE summary line
- Group similar changes: "Update import path" x20 → "- Migrate module paths"
- NEVER list individual files in body
- ONLY describe what is in input - NEVER invent features
- Always output: header + blank line + body lines

Output the commit message directly, no explanation.`

	// LLMCommitPromptZH is the system prompt for generating commit messages in Chinese.
	LLMCommitPromptZH = `生成一条 Angular 格式的 commit 消息.

格式 (必须严格遵循):
<type>(<scope>): <中文描述>

- <正文第1行>
- <正文第2行, 如需要>

规则:
- type: feat|fix|docs|refactor|test|chore|perf
- scope: 主要改动的组件 (llm, config, ui, git, api)
- subject: 这个提交做了什么, 最多 50 字, 不加句号
- body: 最多 1-3 行, 每行以 "- " 开头

重要:
- 正文是必须的 - 至少一行以 "- " 开头
- 归纳总结, 不要逐一列举 - 20 个文件同样的改动只写一行总结
- 合并相似变更: "Update import path" x20 → "- 迁移模块路径"
- 正文中不要列出单个文件名
- 只描述输入中的内容 - 不要编造功能
- 始终输出: 标题行 + 空行 + 正文行

直接输出 commit 消息, 不要解释.`

	// LLMCommitPromptBilingual is the system prompt for generating bilingual commit messages.
	LLMCommitPromptBilingual = `Generate ONE commit message in Angular format with bilingual subject.

FORMAT (MUST follow exactly):
<type>(<scope>): <english> (<中文>)

- <中文正文第1行>
- <中文正文第2行, 如需要>

RULES:
- type: feat|fix|docs|refactor|test|chore|perf
- scope: main component (llm, config, ui, git, api)
- subject: "english (中文)" format, no period
- body: 1-3 lines MAX in Chinese, each starting with "- "

CRITICAL:
- BODY IS MANDATORY - at least one line starting with "- "
- SUMMARIZE, don't enumerate - if 20 files have same change, write ONE summary line
- Group similar changes: "Update import path" x20 → "- 迁移模块路径"
- NEVER list individual files in body
- ONLY describe what is in input - NEVER invent features
- Always output: header + blank line + body lines

Output the commit message directly, no explanation.`
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
