// Package config defines constants and configuration for gitflow-toolkit.
package config

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

// CommitTypeNames returns all commit type names.
func CommitTypeNames() []string {
	names := make([]string, len(CommitTypes))
	for i, ct := range CommitTypes {
		names[i] = ct.Name
	}
	return names
}

// CommitMessagePattern is the regex pattern for validating commit messages.
const CommitMessagePattern = `^(feat|fix|docs|style|refactor|test|chore|perf|hotfix)\((\S.*)\):\s(\S.*)|^Merge.*`

// Environment variable names.
const (
	StrictHostKeyEnv = "GIT_SSH_STRICT_HOST_KEY"
)

// Binary and path constants.
const (
	// BinaryName is the name of the main binary.
	BinaryName = "gitflow-toolkit"

	// GitCommandPrefix is the prefix for git subcommands.
	GitCommandPrefix = "git-"

	// ToolkitHomeDir is the name of the toolkit home directory.
	ToolkitHomeDir = ".gitflow-toolkit"

	// HooksDir is the name of the hooks directory.
	HooksDir = "hooks"

	// CommitMsgHook is the name of the commit-msg hook.
	CommitMsgHook = "commit-msg"

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
