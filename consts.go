package main

type CommitType string

const (
	FEAT     CommitType = "feat"
	FIX      CommitType = "fix"
	DOCS     CommitType = "docs"
	STYLE    CommitType = "style"
	REFACTOR CommitType = "refactor"
	TEST     CommitType = "test"
	CHORE    CommitType = "chore"
	PERF     CommitType = "perf"
	HOTFIX   CommitType = "hotfix"
)

const commitMessagePattern = `^(feat|fix|docs|style|refactor|test|chore|perf|hotfix)\((\S.*)\):\s(\S.*)|^Merge.*`

const commitBodyEditPattern = `^\/\/\s*(?i)edit`

const commitMessageTpl = `{{ .Type }}({{ .Scope }}): {{ .Subject }}

{{ .Body }}

{{ .Footer }}

{{ .Sob }}
`

const commitMessageCheckFailedTpl = `
######################################################
##                                                  ##
##    💔 The commit message is not standardized.    ##
##    💔 It must match the regular expression:      ##
##                                                  ##
##    ^(feat|fix|docs|style|refactor|test|chore|    ##
##     perf|hotfix)\((\S.*)\):\s(\S.*)|^Merge.*     ##
##                                                  ##
######################################################`
