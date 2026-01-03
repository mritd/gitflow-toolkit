package git

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/mritd/gitflow-toolkit/v3/internal/config"
)

// ErrInvalidCommitMessage is returned when a commit message doesn't match the pattern.
var ErrInvalidCommitMessage = errors.New("commit message does not follow conventional format")

// ValidateCommitMessage validates a commit message file against the pattern.
func ValidateCommitMessage(filepath string) error {
	bs, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read commit message file: %w", err)
	}

	return ValidateCommitMessageContent(string(bs))
}

// ValidateCommitMessageContent validates commit message content against the pattern.
func ValidateCommitMessageContent(content string) error {
	reg := regexp.MustCompile(config.CommitMessagePattern)
	matches := reg.FindStringSubmatch(content)

	if len(matches) != 4 {
		return ErrInvalidCommitMessage
	}

	return nil
}
