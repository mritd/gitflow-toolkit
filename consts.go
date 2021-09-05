package main

const (
	FEAT     string = "feat"
	FIX      string = "fix"
	DOCS     string = "docs"
	STYLE    string = "style"
	REFACTOR string = "refactor"
	TEST     string = "test"
	CHORE    string = "chore"
	PERF     string = "perf"
	HOTFIX   string = "hotfix"
)

const (
	FEAT_DESC     string = "FEAT (Introducing new features)"
	FIX_DESC      string = "FIX (Bug fix)"
	DOCS_DESC     string = "DOCS (Writing docs)"
	STYLE_DESC    string = "STYLE (Improving structure/format of the code)"
	REFACTOR_DESC string = "REFACTOR (Refactoring code)"
	TEST_DESC     string = "TEST (When adding missing tests)"
	CHORE_DESC    string = "CHORE (Changing CI/CD)"
	PERF_DESC     string = "PERF (Improving performance)"
	HOTFIX_DESC   string = "HOTFIX (Bug fix urgently)"
)

const (
	UI_SUCCESS_TITLE = `🟢 COMMIT SUCCESS`
	UI_FAILED_TITLE  = `🔴 COMMIT FAILED`
	UI_SUCCESS_MSG   = `Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.`
)

const commitMessagePattern = `^(feat|fix|docs|style|refactor|test|chore|perf|hotfix)\((\S.*)\):\s(\S.*)|^Merge.*`

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
