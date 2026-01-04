package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/config"
)

// Lucky commit errors.
var (
	ErrLuckyPrefixEmpty    = errors.New("invalid lucky commit prefix: cannot be empty")
	ErrLuckyPrefixTooLong  = fmt.Errorf("invalid lucky commit prefix: maximum length is %d characters", consts.LuckyCommitMaxLen)
	ErrLuckyPrefixInvalid  = errors.New("invalid lucky commit prefix: must contain only hex characters [0-9a-f]")
	ErrLuckyCommitNotFound = fmt.Errorf("%s not found in PATH, install it from: %s", consts.LuckyCommitBinary, consts.LuckyCommitURL)
)

// GetLuckyPrefix returns the lucky commit prefix from gitconfig.
// Returns empty string if not set.
func GetLuckyPrefix() string {
	return config.GetString(config.GitConfigLuckyCommitPrefix, "")
}

// ValidateLuckyPrefix validates and normalizes the prefix.
// Returns lowercase prefix or error if invalid.
func ValidateLuckyPrefix(prefix string) (string, error) {
	if prefix == "" {
		return "", ErrLuckyPrefixEmpty
	}

	if len(prefix) > consts.LuckyCommitMaxLen {
		return "", ErrLuckyPrefixTooLong
	}

	prefix = strings.ToLower(prefix)
	for _, c := range prefix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", ErrLuckyPrefixInvalid
		}
	}

	return prefix, nil
}

// CheckLuckyCommit checks if lucky_commit executable exists in PATH.
// Returns nil if found, error with download instructions if not.
func CheckLuckyCommit() error {
	_, err := exec.LookPath(consts.LuckyCommitBinary)
	if err != nil {
		return ErrLuckyCommitNotFound
	}
	return nil
}

// LuckyCommitCmd creates an exec.Cmd for running lucky_commit with the given prefix.
func LuckyCommitCmd(prefix string) *exec.Cmd {
	return exec.Command(consts.LuckyCommitBinary, prefix)
}

// GetHeadHash returns the current HEAD commit hash.
func GetHeadHash() (string, error) {
	return Run("rev-parse", "HEAD")
}
