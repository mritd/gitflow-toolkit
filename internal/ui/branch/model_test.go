package branch

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

func TestNewModel(t *testing.T) {
	m := NewModel("feat", "my-feature")

	if m.branchType != "feat" {
		t.Errorf("branchType = %q, want 'feat'", m.branchType)
	}

	if m.branchName != "my-feature" {
		t.Errorf("branchName = %q, want 'my-feature'", m.branchName)
	}

	if m.fullName != "feat/my-feature" {
		t.Errorf("fullName = %q, want 'feat/my-feature'", m.fullName)
	}

	if m.state != StateCreating {
		t.Errorf("state = %v, want StateCreating", m.state)
	}
}

func TestModel_Update_BranchDone_Success(t *testing.T) {
	m := NewModel("feat", "test")

	msg := branchDoneMsg{result: "Switched to branch", err: nil}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.state != StateSuccess {
		t.Errorf("state = %v, want StateSuccess", model.state)
	}

	if model.result != "Switched to branch" {
		t.Errorf("result = %q, want 'Switched to branch'", model.result)
	}

	if model.err != nil {
		t.Errorf("err = %v, want nil", model.err)
	}
}

func TestModel_Update_BranchDone_Failed(t *testing.T) {
	m := NewModel("feat", "test")

	testErr := errors.New("branch already exists")
	msg := branchDoneMsg{result: "", err: testErr}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.state != StateFailed {
		t.Errorf("state = %v, want StateFailed", model.state)
	}

	if model.err != testErr {
		t.Errorf("err = %v, want %v", model.err, testErr)
	}
}

func TestModel_Update_KeyMsg(t *testing.T) {
	tests := []struct {
		key      string
		wantQuit bool
	}{
		{"ctrl+c", true},
		{"q", true},
		{"esc", true},
		{"enter", true},
		{"a", false},
		{"up", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := NewModel("feat", "test")
			m.state = StateSuccess // Ensure we're in a state that can quit

			var msg tea.KeyMsg
			switch tt.key {
			case "ctrl+c":
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			case "enter":
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			_, cmd := m.Update(msg)

			gotQuit := cmd != nil
			if gotQuit != tt.wantQuit {
				t.Errorf("Update(%q) quit = %v, want %v", tt.key, gotQuit, tt.wantQuit)
			}
		})
	}
}

func TestModel_View_Creating(t *testing.T) {
	m := NewModel("feat", "my-feature")
	m.state = StateCreating

	view := m.View()

	if !strings.Contains(view, "feat/my-feature") {
		t.Error("View should contain branch name")
	}

	if !strings.Contains(view, "Creating") {
		t.Error("View should contain 'Creating'")
	}
}

func TestModel_View_Success(t *testing.T) {
	m := NewModel("fix", "bug-123")
	m.state = StateSuccess
	m.result = "Switched to branch 'fix/bug-123'"

	view := m.View()

	if !strings.Contains(view, common.SymbolSuccess) {
		t.Error("View should contain success symbol")
	}

	if !strings.Contains(view, "fix/bug-123") {
		t.Error("View should contain branch name")
	}

	if !strings.Contains(view, "Branch created") {
		t.Error("View should contain 'Branch created'")
	}

	if !strings.Contains(view, "Switched to branch") {
		t.Error("View should contain result message")
	}
}

func TestModel_View_Failed(t *testing.T) {
	m := NewModel("feat", "existing")
	m.state = StateFailed
	m.err = errors.New("branch 'feat/existing' already exists")

	view := m.View()

	if !strings.Contains(view, common.SymbolError) {
		t.Error("View should contain error symbol")
	}

	if !strings.Contains(view, "failed") {
		t.Error("View should contain 'failed'")
	}

	if !strings.Contains(view, "already exists") {
		t.Error("View should contain error message")
	}
}

func TestModel_Error(t *testing.T) {
	m := NewModel("feat", "test")

	if m.Error() != nil {
		t.Error("Error() should be nil initially")
	}

	testErr := errors.New("test error")
	m.err = testErr

	if m.Error() != testErr {
		t.Errorf("Error() = %v, want %v", m.Error(), testErr)
	}
}

func TestModel_IsSuccess(t *testing.T) {
	m := NewModel("feat", "test")

	if m.IsSuccess() {
		t.Error("IsSuccess() should be false initially")
	}

	m.state = StateSuccess
	if !m.IsSuccess() {
		t.Error("IsSuccess() should be true after success")
	}

	m.state = StateFailed
	if m.IsSuccess() {
		t.Error("IsSuccess() should be false after failure")
	}
}

func TestState_Values(t *testing.T) {
	// Verify state constants
	if StateCreating != 0 {
		t.Errorf("StateCreating = %d, want 0", StateCreating)
	}
	if StateSuccess != 1 {
		t.Errorf("StateSuccess = %d, want 1", StateSuccess)
	}
	if StateFailed != 2 {
		t.Errorf("StateFailed = %d, want 2", StateFailed)
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel("feat", "test")
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}
