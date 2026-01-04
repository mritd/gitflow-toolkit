package git

import (
	"fmt"
	"strings"

	"github.com/mritd/gitflow-toolkit/v3/consts"
)

// branchAliases maps branch prefixes to commit types.
var branchAliases = map[string]string{
	// feat
	"feat":    consts.Feat,
	"feature": consts.Feat,
	// fix
	"fix":    consts.Fix,
	"bugfix": consts.Fix,
	"bug":    consts.Fix,
	// docs
	"docs":     consts.Docs,
	"doc":      consts.Docs,
	"document": consts.Docs,
	// style
	"style": consts.Style,
	// refactor
	"refactor": consts.Refactor,
	"refact":   consts.Refactor,
	// test
	"test":    consts.Test,
	"testing": consts.Test,
	// chore
	"chore": consts.Chore,
	// perf
	"perf":        consts.Perf,
	"performance": consts.Perf,
	// hotfix
	"hotfix": consts.Hotfix,
}

// ParseBranchType extracts the commit type from a branch name.
// Supports formats: type/name, type-name, type_name
// Returns empty string if no match found.
func ParseBranchType(branch string) string {
	if branch == "" {
		return ""
	}

	// Find the prefix before /, -, or _
	var prefix string
	for i, r := range branch {
		if r == '/' || r == '-' || r == '_' {
			prefix = branch[:i]
			break
		}
	}

	// If no separator found, the whole branch name might be the type
	if prefix == "" {
		prefix = branch
	}

	prefix = strings.ToLower(prefix)
	if commitType, ok := branchAliases[prefix]; ok {
		return commitType
	}

	return ""
}

// CreateBranch creates a new branch with the given name.
func CreateBranch(name string) (string, error) {
	if err := RepoCheck(); err != nil {
		return "", err
	}
	return Run("switch", "-c", name)
}

// CreateTypedBranch creates a new branch with type prefix (e.g., feat/name).
func CreateTypedBranch(commitType, name string) (string, error) {
	branchName := fmt.Sprintf("%s/%s", commitType, name)
	return CreateBranch(branchName)
}

// Push pushes the current branch to origin.
func Push() (string, error) {
	if err := RepoCheck(); err != nil {
		return "", err
	}

	branch, err := CurrentBranch()
	if err != nil {
		return "", err
	}

	msg, err := Run("push", "origin", branch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Push to origin/%s success.\n\n%s", branch, msg), nil
}
