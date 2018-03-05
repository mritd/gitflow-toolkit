package config

type CommitType string

const (
	COMMIT_TYPE_FEAT CommitType = "feat"
	COMMIT_TYPE_FIX CommitType = "fix"
	COMMIT_TYPE_HOTFIX CommitType = "hotfix"
	COMMIT_TYPE_DOCS CommitType = "docs"
	COMMIT_TYPE_STYLE CommitType = "style"
	COMMIT_TYPE_REFACTOR CommitType = "refactor"
	COMMIT_TYPE_TEST CommitType = "test"
	COMMIT_TYPE_CHORE CommitType = "chore"
	COMMIT_TYPE_PERF CommitType = "perf"
)

type CommitMessage struct{
	Type string
	Scope string
	Subject string
	Body string
	Footer string
}