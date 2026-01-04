package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/mritd/gitflow-toolkit/v3/consts"
)

// CommitMessage represents a structured commit message.
type CommitMessage struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
	SOB     string // Signed-off-by line
}

// String formats the commit message.
func (m CommitMessage) String() string {
	var sb strings.Builder

	// Header: type(scope): subject
	sb.WriteString(fmt.Sprintf("%s(%s): %s", m.Type, m.Scope, m.Subject))

	// Body (if present)
	if m.Body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(m.Body)
	}

	// Footer (if present)
	if m.Footer != "" {
		sb.WriteString("\n\n")
		sb.WriteString(m.Footer)
	}

	// Signed-off-by
	if m.SOB != "" {
		sb.WriteString("\n\n")
		sb.WriteString(m.SOB)
	}

	sb.WriteString("\n")
	return sb.String()
}

// Commit creates a new commit with the given message.
func Commit(msg CommitMessage) error {
	if err := HasStagedFiles(); err != nil {
		return err
	}

	f, err := os.CreateTemp("", consts.TempFilePrefix+"-commit")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = f.WriteString(msg.String())
	if err != nil {
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	_, err = Run("commit", "-F", f.Name())
	return err
}

// CreateSOB creates a Signed-off-by line.
func CreateSOB() string {
	name, email := Author()
	if name == "" || email == "" {
		return ""
	}
	return fmt.Sprintf("Signed-off-by: %s <%s>", name, email)
}
