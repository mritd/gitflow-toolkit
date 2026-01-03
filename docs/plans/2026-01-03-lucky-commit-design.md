# Lucky Commit Feature Design

## Overview

Add lucky commit support to gitflow-toolkit. When `GITFLOW_LUCKY_COMMIT` environment variable is set with a hex prefix, the commit hash will be brute-forced to start with that prefix using the external `lucky_commit` tool.

## Flow

```
Start git ci
    │
    ▼
Check GITFLOW_LUCKY_COMMIT env var
    │
    ├─ Not set → Normal commit flow
    │
    └─ Set → Validate prefix format (hex, ≤16 chars, lowercase)
                │
                ├─ Invalid format → Error, exit
                │
                └─ Valid format → Check lucky_commit executable
                                    │
                                    ├─ Not found → Show download prompt, abort
                                    │
                                    └─ Found → Continue normal commit flow
                                                │
                                                ▼
                                          Commit success
                                                │
                                                ▼
                                    Show spinner animation
                                    Execute lucky_commit <prefix>
                                                │
                                    ├─ Success → Show final lucky hash
                                    ├─ Ctrl+C → Warning, preserve original commit
                                    └─ Failure → Warning, preserve original commit
```

## Code Structure

### New/Modified Files

```
internal/
  git/
    lucky.go          # Lucky commit core logic
                      # - ValidateLuckyPrefix(prefix string) (string, error)
                      # - CheckLuckyCommit() error  
                      # - RunLuckyCommit(prefix string) error

  ui/
    commit/
      model.go        # Modified: trigger lucky commit after success

    common/
      spinner.go      # New: interruptible spinner component

cmd/
  commit.go           # Modified: check env var and executable at startup
```

### Key Function Signatures

```go
// internal/git/lucky.go

// ValidateLuckyPrefix validates and normalizes the prefix.
// Returns lowercase prefix or error if invalid.
func ValidateLuckyPrefix(prefix string) (string, error)

// CheckLuckyCommit checks if lucky_commit executable exists in PATH.
// Returns nil if found, error with download instructions if not.
func CheckLuckyCommit() error

// RunLuckyCommit runs lucky_commit with the given prefix.
// Modifies HEAD commit to have hash starting with prefix.
func RunLuckyCommit(prefix string) error
```

## Spinner Component

### Behavior

```go
// internal/ui/common/spinner.go

type SpinnerModel struct {
    title    string           // Display title
    cmd      *exec.Cmd        // External process
    done     bool             // Completion flag
    err      error            // Execution result
    canceled bool             // User Ctrl+C flag
}

// Usage
func RunWithSpinner(title string, cmd *exec.Cmd) (canceled bool, err error)
```

### Display Style

```
Lucky Commit

   ████████████░░░░░░░░  Searching: abc123...

   Current: 1a2b3c4 → abc1f2e → abc12d8 → abc123a

Press Ctrl+C to skip
```

### Success State

```
Lucky Commit

   ████████████████████  Found!

   Result: abc123f8b2e1a4c5d6e7f8a9b0c1d2e3f4a5b6c7

Press any key to continue
```

### Visual Elements

- Title: Use existing `TitleStyle` (with background color)
- Progress bar: Gradient pulse animation
- Hash scroll: Refresh every 100ms, random hashes matching prefix
- Footer hint: Gray text, `PaddingTop(1)`
- On success: Full green progress bar, display real hash

### Interrupt Handling

```
User presses Ctrl+C
    │
    ▼
Send SIGTERM to lucky_commit process
    │
    ▼
Wait for process exit (brief timeout)
    │
    ▼
Return canceled=true, preserve original commit
```

## Error Handling

| Scenario | Type | Message |
|----------|------|---------|
| Non-hex characters in prefix | Error | `invalid lucky commit prefix: must contain only hex characters [0-9a-f]` |
| Prefix exceeds 16 characters | Error | `invalid lucky commit prefix: maximum length is 16 characters` |
| Empty prefix | Error | `invalid lucky commit prefix: cannot be empty` |
| `lucky_commit` not found | Error | `lucky_commit not found in PATH, install it from: https://github.com/not-an-aardvark/lucky-commit` |
| User Ctrl+C interrupt | Warning | `lucky commit skipped, original commit preserved` |
| `lucky_commit` execution failed | Warning | `lucky commit failed: <error>, original commit preserved` |
| Success | Success | Show commit result with lucky hash |

Note: Ctrl+C and execution failures are warnings (not errors) since the original commit is preserved.

## Configuration

| Item | Value |
|------|-------|
| Environment variable | `GITFLOW_LUCKY_COMMIT` |
| Max prefix length | 16 characters |
| Allowed characters | `[0-9a-f]` (auto-lowercase) |
| External dependency | `lucky_commit` from https://github.com/not-an-aardvark/lucky-commit |

---

# Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add lucky commit support that brute-forces commit hash to start with a specified prefix.

**Architecture:** Environment variable triggers validation at startup, commit flow proceeds normally, then lucky_commit runs with animated spinner UI.

**Tech Stack:** Go, Bubble Tea, bubbles/spinner, os/exec

---

## Task 1: Add Config Constants

**Files:**
- Modify: `internal/config/config.go:51-56`
- Test: `internal/config/config_test.go`

**Step 1: Add constants**

```go
// Environment variable names.
const (
	// StrictHostKeyEnv controls SSH strict host key checking.
	// Set to "true" to enable strict host key checking (default: disabled).
	StrictHostKeyEnv = "GITFLOW_SSH_STRICT_HOST_KEY"

	// LuckyCommitEnv is the environment variable for lucky commit prefix.
	// When set, commit hash will be brute-forced to start with this prefix.
	LuckyCommitEnv = "GITFLOW_LUCKY_COMMIT"

	// LuckyCommitBinary is the name of the lucky_commit executable.
	LuckyCommitBinary = "lucky_commit"

	// LuckyCommitMaxLen is the maximum length of lucky commit prefix.
	LuckyCommitMaxLen = 8

	// LuckyCommitURL is the download URL for lucky_commit.
	LuckyCommitURL = "https://github.com/not-an-aardvark/lucky-commit"
)
```

**Step 2: Run tests**

Run: `go test -v ./internal/config/...`
Expected: PASS

**Step 3: Commit**

```
feat(config): add lucky commit constants
```

---

## Task 2: Implement Lucky Commit Core Logic

**Files:**
- Create: `internal/git/lucky.go`
- Create: `internal/git/lucky_test.go`

**Step 1: Write failing tests**

```go
package git

import (
	"os"
	"os/exec"
	"testing"
)

func TestValidateLuckyPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		want    string
		wantErr bool
	}{
		{"valid lowercase", "abc123", "abc123", false},
		{"valid uppercase converts", "ABC123", "abc123", false},
		{"valid mixed case", "AbC123", "abc123", false},
		{"empty string", "", "", true},
		{"too long", "123456789", "", true},
		{"exactly 8 chars", "12345678", "12345678", false},
		{"invalid chars", "xyz123", "", true},
		{"invalid with space", "abc 123", "", true},
		{"valid all zeros", "00000000", "00000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateLuckyPrefix(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLuckyPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateLuckyPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckLuckyCommit(t *testing.T) {
	// Save original PATH
	origPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", origPath) }()

	// Test with empty PATH (lucky_commit not found)
	_ = os.Setenv("PATH", "")
	err := CheckLuckyCommit()
	if err == nil {
		t.Error("CheckLuckyCommit() should return error when not in PATH")
	}

	// Test with lucky_commit in PATH (if available)
	_ = os.Setenv("PATH", origPath)
	if _, lookErr := exec.LookPath("lucky_commit"); lookErr == nil {
		err = CheckLuckyCommit()
		if err != nil {
			t.Errorf("CheckLuckyCommit() unexpected error: %v", err)
		}
	}
}

func TestGetLuckyPrefix(t *testing.T) {
	// Save original env
	origVal := os.Getenv("GITFLOW_LUCKY_COMMIT")
	defer func() { _ = os.Setenv("GITFLOW_LUCKY_COMMIT", origVal) }()

	// Test not set
	_ = os.Unsetenv("GITFLOW_LUCKY_COMMIT")
	prefix := GetLuckyPrefix()
	if prefix != "" {
		t.Errorf("GetLuckyPrefix() = %v, want empty", prefix)
	}

	// Test set
	_ = os.Setenv("GITFLOW_LUCKY_COMMIT", "abc123")
	prefix = GetLuckyPrefix()
	if prefix != "abc123" {
		t.Errorf("GetLuckyPrefix() = %v, want abc123", prefix)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/git/... -run "Lucky|Prefix"`
Expected: FAIL with "undefined: ValidateLuckyPrefix"

**Step 3: Implement the functions**

```go
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mritd/gitflow-toolkit/v3/internal/config"
)

// Lucky commit errors.
var (
	ErrLuckyPrefixEmpty   = errors.New("invalid lucky commit prefix: cannot be empty")
	ErrLuckyPrefixTooLong = fmt.Errorf("invalid lucky commit prefix: maximum length is %d characters", config.LuckyCommitMaxLen)
	ErrLuckyPrefixInvalid = errors.New("invalid lucky commit prefix: must contain only hex characters [0-9a-f]")
	ErrLuckyCommitNotFound = fmt.Errorf("%s not found in PATH, install it from: %s", config.LuckyCommitBinary, config.LuckyCommitURL)
)

// GetLuckyPrefix returns the lucky commit prefix from environment variable.
// Returns empty string if not set.
func GetLuckyPrefix() string {
	return os.Getenv(config.LuckyCommitEnv)
}

// ValidateLuckyPrefix validates and normalizes the prefix.
// Returns lowercase prefix or error if invalid.
func ValidateLuckyPrefix(prefix string) (string, error) {
	if prefix == "" {
		return "", ErrLuckyPrefixEmpty
	}

	if len(prefix) > config.LuckyCommitMaxLen {
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
	_, err := exec.LookPath(config.LuckyCommitBinary)
	if err != nil {
		return ErrLuckyCommitNotFound
	}
	return nil
}

// LuckyCommitCmd creates an exec.Cmd for running lucky_commit with the given prefix.
func LuckyCommitCmd(prefix string) *exec.Cmd {
	return exec.Command(config.LuckyCommitBinary, prefix)
}

// GetHeadHash returns the current HEAD commit hash.
func GetHeadHash() (string, error) {
	return Run("rev-parse", "HEAD")
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/git/... -run "Lucky|Prefix"`
Expected: PASS

**Step 5: Commit**

```
feat(git): add lucky commit validation and check functions
```

---

## Task 3: Implement Lucky Spinner UI Component

**Files:**
- Create: `internal/ui/common/lucky.go`
- Create: `internal/ui/common/lucky_test.go`

**Step 1: Write the lucky spinner component**

```go
package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LuckyResult represents the result of a lucky commit operation.
type LuckyResult struct {
	Cancelled bool
	Err       error
	Hash      string
}

// luckyModel is the bubbletea model for lucky commit spinner.
type luckyModel struct {
	prefix      string
	cmd         *exec.Cmd
	spinner     spinner.Model
	progressPos int
	fakeHash    string
	done        bool
	cancelled   bool
	err         error
	hash        string
	getHash     func() (string, error)
}

// luckyDoneMsg is sent when lucky_commit completes.
type luckyDoneMsg struct {
	err error
}

// luckyTickMsg is sent for animation updates.
type luckyTickMsg struct{}

// newLuckyModel creates a new lucky commit model.
func newLuckyModel(prefix string, cmd *exec.Cmd, getHash func() (string, error)) luckyModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return luckyModel{
		prefix:      prefix,
		cmd:         cmd,
		spinner:     s,
		progressPos: 0,
		fakeHash:    generateFakeHash(prefix),
		getHash:     getHash,
	}
}

// generateFakeHash generates a random hash starting with the given prefix.
func generateFakeHash(prefix string) string {
	remaining := 40 - len(prefix)
	if remaining <= 0 {
		return prefix
	}
	
	randomBytes := make([]byte, remaining/2+1)
	_, _ = rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	
	return prefix + randomHex[:remaining]
}

func (m luckyModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runLuckyCommit(),
		m.tickAnimation(),
	)
}

func (m luckyModel) runLuckyCommit() tea.Cmd {
	return func() tea.Msg {
		err := m.cmd.Run()
		return luckyDoneMsg{err: err}
	}
}

func (m luckyModel) tickAnimation() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return luckyTickMsg{}
	})
}

func (m luckyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			// Send SIGTERM to the process
			if m.cmd.Process != nil {
				_ = m.cmd.Process.Signal(syscall.SIGTERM)
			}
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case luckyTickMsg:
		if !m.done {
			m.progressPos = (m.progressPos + 1) % 20
			m.fakeHash = generateFakeHash(m.prefix)
			return m, m.tickAnimation()
		}

	case luckyDoneMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
		} else {
			// Get the actual hash
			if hash, err := m.getHash(); err == nil {
				m.hash = hash
			}
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m luckyModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var sb strings.Builder

	// Title
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorTitleFg).
		Background(ColorTitleBg).
		Bold(true).
		Padding(0, 1)
	sb.WriteString(titleLayout.Render(titleStyle.Render("Lucky Commit")))
	sb.WriteString("\n")

	// Progress bar
	contentLayout := lipgloss.NewStyle().PaddingLeft(4)
	progressBar := m.renderProgressBar()
	sb.WriteString(contentLayout.Render(progressBar + "  Searching: " + m.prefix + "..."))
	sb.WriteString("\n\n")

	// Current hash animation
	hashStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	sb.WriteString(contentLayout.Render("Current: " + hashStyle.Render(m.fakeHash)))
	sb.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)
	sb.WriteString(helpStyle.Render("Press Ctrl+C to skip"))
	sb.WriteString("\n")

	return sb.String()
}

func (m luckyModel) renderProgressBar() string {
	width := 20
	filledStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
	emptyStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	var bar strings.Builder
	pulseWidth := 6
	pulseStart := m.progressPos

	for i := 0; i < width; i++ {
		// Create pulse effect
		inPulse := false
		for j := 0; j < pulseWidth; j++ {
			if (pulseStart+j)%width == i {
				inPulse = true
				break
			}
		}
		if inPulse {
			bar.WriteString(filledStyle.Render("█"))
		} else {
			bar.WriteString(emptyStyle.Render("░"))
		}
	}

	return bar.String()
}

// RunLuckyCommit runs lucky_commit with an animated spinner.
// Returns the result of the operation.
func RunLuckyCommit(prefix string, cmd *exec.Cmd, getHash func() (string, error)) LuckyResult {
	m := newLuckyModel(prefix, cmd, getHash)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return LuckyResult{Err: err}
	}

	result := finalModel.(luckyModel)
	return LuckyResult{
		Cancelled: result.cancelled,
		Err:       result.err,
		Hash:      result.hash,
	}
}
```

**Step 2: Write basic test**

```go
package common

import (
	"strings"
	"testing"
)

func TestGenerateFakeHash(t *testing.T) {
	tests := []struct {
		prefix string
		length int
	}{
		{"abc", 40},
		{"12345678", 40},
		{"a", 40},
		{"", 40},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			hash := generateFakeHash(tt.prefix)
			if len(hash) != tt.length {
				t.Errorf("generateFakeHash(%q) length = %d, want %d", tt.prefix, len(hash), tt.length)
			}
			if !strings.HasPrefix(hash, tt.prefix) {
				t.Errorf("generateFakeHash(%q) = %q, should start with %q", tt.prefix, hash, tt.prefix)
			}
		})
	}
}
```

**Step 3: Run tests**

Run: `go test -v ./internal/ui/common/... -run "FakeHash"`
Expected: PASS

**Step 4: Commit**

```
feat(ui): add lucky commit spinner component
```

---

## Task 4: Integrate Lucky Commit into Commit Flow

**Files:**
- Modify: `internal/ui/commit/model.go`
- Modify: `cmd/commit.go`

**Step 1: Update Result struct and Run function**

In `internal/ui/commit/model.go`, update the Result struct and Run function:

```go
// Result represents the result of the commit flow.
type Result struct {
	Cancelled    bool
	Err          error
	Message      git.CommitMessage
	LuckySkipped bool   // true if lucky commit was skipped (Ctrl+C)
	LuckyFailed  error  // error if lucky commit failed
	Hash         string // final commit hash (may be lucky hash)
}

// Run runs the interactive commit flow.
// The luckyPrefix parameter is the validated lucky commit prefix (empty if not enabled).
// Returns the result of the commit operation.
func Run(luckyPrefix string) Result {
	var result Result

	// Step 1: Select commit type
	commitType, err := runSelector()
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Step 2: Input all fields (scope, subject, body, footer)
	inputs, err := runInputs(commitType)
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Create SOB
	sob := git.CreateSOB()

	// If body is empty, use subject as body
	body := inputs.body
	if body == "" {
		body = inputs.subject
	}

	// Build commit message
	result.Message = git.CommitMessage{
		Type:    commitType,
		Scope:   inputs.scope,
		Subject: inputs.subject,
		Body:    body,
		Footer:  inputs.footer,
		SOB:     sob,
	}

	// Step 3: Confirm and commit
	confirmed, err := confirmCommit(result.Message)
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	if !confirmed {
		result.Cancelled = true
		return result
	}

	// Perform commit
	if err := git.Commit(result.Message); err != nil {
		result.Err = err
		return result
	}

	// Run lucky commit if prefix is set
	if luckyPrefix != "" {
		cmd := git.LuckyCommitCmd(luckyPrefix)
		luckyResult := common.RunLuckyCommit(luckyPrefix, cmd, git.GetHeadHash)
		
		if luckyResult.Cancelled {
			result.LuckySkipped = true
		} else if luckyResult.Err != nil {
			result.LuckyFailed = luckyResult.Err
		}
		result.Hash = luckyResult.Hash
	}

	// Get hash if not already set
	if result.Hash == "" {
		result.Hash, _ = git.GetHeadHash()
	}

	return result
}
```

**Step 2: Update cmd/commit.go**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v3/internal/config"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/commit"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// commitCmd represents the commit command.
var commitCmd = &cobra.Command{
	Use:     config.CmdCommit,
	Aliases: []string{"commit"},
	Short:   "Interactive commit with conventional commit format",
	Long: `Create a commit following the conventional commit format.

This command provides an interactive TUI to help you create
properly formatted commit messages with type, scope, subject,
body, and footer.

The commit message format follows the Angular specification:
  type(scope): subject

  body

  footer

  Signed-off-by: Name <email>`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Check if there are staged files
	if err := git.HasStagedFiles(); err != nil {
		return renderError(cmd, "No staged files", err)
	}

	// Check lucky commit configuration at startup
	var luckyPrefix string
	if rawPrefix := git.GetLuckyPrefix(); rawPrefix != "" {
		// Validate prefix format
		prefix, err := git.ValidateLuckyPrefix(rawPrefix)
		if err != nil {
			return renderError(cmd, "Lucky commit", err)
		}

		// Check lucky_commit executable exists
		if err := git.CheckLuckyCommit(); err != nil {
			return renderError(cmd, "Lucky commit", err)
		}

		luckyPrefix = prefix
	}

	// Run the interactive commit flow
	result := commit.Run(luckyPrefix)

	if result.Cancelled {
		r := common.Warning("Commit cancelled", "Operation was cancelled by user.")
		fmt.Print(common.RenderResult(r))
		return nil
	}

	if result.Err != nil {
		return renderError(cmd, "Commit failed", result.Err)
	}

	// Build success message
	msg := result.Message
	content := common.FormatCommitMessage(common.CommitMessageContent{
		Type:    msg.Type,
		Scope:   msg.Scope,
		Subject: msg.Subject,
		Body:    msg.Body,
		Footer:  msg.Footer,
		SOB:     msg.SOB,
	})

	// Add hash info
	if result.Hash != "" {
		content += "\n\n" + common.StyleMuted.Render("Hash: "+result.Hash)
	}

	// Handle lucky commit results
	if result.LuckySkipped {
		r := common.Warning("Commit created (lucky skipped)", content)
		r.Note = "lucky commit skipped, original commit preserved"
		fmt.Print(common.RenderResult(r))
		return nil
	}

	if result.LuckyFailed != nil {
		r := common.Warning("Commit created (lucky failed)", content)
		r.Note = fmt.Sprintf("lucky commit failed: %s, original commit preserved", result.LuckyFailed)
		fmt.Print(common.RenderResult(r))
		return nil
	}

	r := common.Success("Commit created", content)
	r.Note = "Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live."
	fmt.Print(common.RenderResult(r))
	return nil
}
```

**Step 3: Run full test suite**

Run: `go test -v ./...`
Expected: PASS

**Step 4: Commit**

```
feat(commit): integrate lucky commit into commit flow
```

---

## Task 5: Manual Testing

**Step 1: Build the binary**

Run: `go build -o gitflow-toolkit .`

**Step 2: Test without lucky commit**

Run: `./gitflow-toolkit ci`
Expected: Normal commit flow, no lucky commit

**Step 3: Test with invalid prefix**

Run: `GITFLOW_LUCKY_COMMIT=xyz ./gitflow-toolkit ci`
Expected: Error "must contain only hex characters"

**Step 4: Test with prefix too long**

Run: `GITFLOW_LUCKY_COMMIT=123456789 ./gitflow-toolkit ci`
Expected: Error "maximum length is 8 characters"

**Step 5: Test without lucky_commit installed**

Run: `GITFLOW_LUCKY_COMMIT=abc PATH= ./gitflow-toolkit ci`
Expected: Error "lucky_commit not found in PATH"

**Step 6: Test with valid config (if lucky_commit installed)**

Run: `GITFLOW_LUCKY_COMMIT=abc ./gitflow-toolkit ci`
Expected: Spinner animation, then commit with hash starting with "abc"

---

## Task 6: Update Documentation

**Files:**
- Modify: `README.md`

**Step 1: Add lucky commit section to README**

Add after Environment Variables section:

```markdown
### Lucky Commit

Generate commit hashes with a specific prefix using [lucky_commit](https://github.com/not-an-aardvark/lucky-commit):

```bash
# Set the desired prefix (hex characters, max 8)
export GITFLOW_LUCKY_COMMIT=abc123

# Commit as usual - hash will start with abc123
git ci
```

Requirements:
- Install `lucky_commit` from https://github.com/not-an-aardvark/lucky-commit
- Prefix must be valid hex characters (0-9, a-f)
- Maximum prefix length is 8 characters
- Press Ctrl+C during search to skip and keep original commit
```

**Step 2: Commit**

```
docs: add lucky commit documentation
```
