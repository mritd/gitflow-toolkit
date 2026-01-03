package push

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.state != StatePushing {
		t.Errorf("state = %v, want StatePushing", m.state)
	}

	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}

	if m.result != "" {
		t.Errorf("result = %q, want empty", m.result)
	}
}

func TestModel_Update_PushDone_Success(t *testing.T) {
	m := NewModel()

	msg := pushDoneMsg{result: "Push to origin/main success.", err: nil}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.state != StateSuccess {
		t.Errorf("state = %v, want StateSuccess", model.state)
	}

	if model.result != "Push to origin/main success." {
		t.Errorf("result = %q, want 'Push to origin/main success.'", model.result)
	}

	if model.err != nil {
		t.Errorf("err = %v, want nil", model.err)
	}
}

func TestModel_Update_PushDone_Failed(t *testing.T) {
	m := NewModel()

	testErr := errors.New("remote rejected")
	msg := pushDoneMsg{result: "", err: testErr}
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
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := NewModel()
			m.state = StateSuccess

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

func TestModel_View_Pushing(t *testing.T) {
	m := NewModel()
	m.state = StatePushing

	view := m.View()

	if !strings.Contains(view, "Pushing") {
		t.Error("View should contain 'Pushing'")
	}
}

func TestModel_View_Success(t *testing.T) {
	m := NewModel()
	m.state = StateSuccess
	m.result = "Push to origin/main success."

	view := m.View()

	if !strings.Contains(view, common.SymbolSuccess) {
		t.Error("View should contain success symbol")
	}

	if !strings.Contains(view, "completed") {
		t.Error("View should contain 'completed'")
	}

	if !strings.Contains(view, "Push to origin/main") {
		t.Error("View should contain result message")
	}
}

func TestModel_View_Failed(t *testing.T) {
	m := NewModel()
	m.state = StateFailed
	m.err = errors.New("remote rejected push")

	view := m.View()

	if !strings.Contains(view, common.SymbolError) {
		t.Error("View should contain error symbol")
	}

	if !strings.Contains(view, "failed") {
		t.Error("View should contain 'failed'")
	}

	if !strings.Contains(view, "remote rejected") {
		t.Error("View should contain error message")
	}
}

func TestModel_Error(t *testing.T) {
	m := NewModel()

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
	m := NewModel()

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
	if StatePushing != 0 {
		t.Errorf("StatePushing = %d, want 0", StatePushing)
	}
	if StateSuccess != 1 {
		t.Errorf("StateSuccess = %d, want 1", StateSuccess)
	}
	if StateFailed != 2 {
		t.Errorf("StateFailed = %d, want 2", StateFailed)
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}
