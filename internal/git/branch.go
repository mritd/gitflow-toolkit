package git

import "fmt"

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
