package consts

type CommitType string

type RepoType string

const GitCmd = "git"

const (
	FEAT     CommitType = "feat"
	FIX      CommitType = "fix"
	DOCS     CommitType = "docs"
	STYLE    CommitType = "style"
	REFACTOR CommitType = "refactor"
	TEST     CommitType = "test"
	CHORE    CommitType = "chore"
	PERF     CommitType = "perf"
	EXIT     CommitType = "exit"
)

const (
	GitHubRepo RepoType = "github"
	GitLabRepo RepoType = "gitlab"
)

const CommitTpl = `{{ .Type }}({{ .Scope }}): {{ .Subject }}

{{ .Body }}

{{ .Footer }}

{{ .Sob }}
`

const CommitMessagePattern = `^(?:fixup!\s*)?(\w*)(\(([\w\$\.\*/-]*)\))?\: (.*)|^Merge\ branch(.*)`
