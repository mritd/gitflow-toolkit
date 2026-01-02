package common

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTaskState(t *testing.T) {
	tests := []struct {
		state    TaskState
		expected int
	}{
		{TaskPending, 0},
		{TaskRunning, 1},
		{TaskSuccess, 2},
		{TaskWarning, 3},
		{TaskFailed, 4},
	}

	for _, tt := range tests {
		if int(tt.state) != tt.expected {
			t.Errorf("TaskState %v = %d, want %d", tt.state, tt.state, tt.expected)
		}
	}
}

func TestWarnErr(t *testing.T) {
	err := WarnErr{Msg: "test warning"}

	if err.Error() != "test warning" {
		t.Errorf("WarnErr.Error() = %q, want %q", err.Error(), "test warning")
	}

	if !IsWarnErr(err) {
		t.Error("IsWarnErr(WarnErr{}) should return true")
	}

	normalErr := errors.New("normal error")
	if IsWarnErr(normalErr) {
		t.Error("IsWarnErr(normal error) should return false")
	}
}

func TestMultiTaskModel_Init(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return nil }},
		{Name: "Task 2", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test", tasks)

	if len(model.tasks) != 2 {
		t.Errorf("len(tasks) = %d, want 2", len(model.tasks))
	}

	if model.currentTask != -1 {
		t.Errorf("currentTask = %d, want -1", model.currentTask)
	}

	if model.done {
		t.Error("done should be false initially")
	}

	if model.title != "Test" {
		t.Errorf("title = %q, want %q", model.title, "Test")
	}
}

func TestMultiTaskModel_Update_TaskDone(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return nil }},
		{Name: "Task 2", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test", tasks)
	model.currentTask = 0
	model.states[0] = TaskRunning

	msg := taskDoneMsg{index: 0, err: nil}
	newModel, _ := model.Update(msg)
	m := newModel.(MultiTaskModel)

	if m.states[0] != TaskSuccess {
		t.Errorf("states[0] = %v, want TaskSuccess", m.states[0])
	}

	if m.currentTask != 1 {
		t.Errorf("currentTask = %d, want 1", m.currentTask)
	}
}

func TestMultiTaskModel_Update_TaskFailed(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return errors.New("failed") }},
		{Name: "Task 2", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test", tasks)
	model.currentTask = 0
	model.states[0] = TaskRunning

	msg := taskDoneMsg{index: 0, err: errors.New("failed")}
	newModel, _ := model.Update(msg)
	m := newModel.(MultiTaskModel)

	if m.states[0] != TaskFailed {
		t.Errorf("states[0] = %v, want TaskFailed", m.states[0])
	}

	if !m.done {
		t.Error("done should be true after failure")
	}
}

func TestMultiTaskModel_Update_TaskWarning(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return WarnErr{Msg: "warning"} }},
		{Name: "Task 2", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test", tasks)
	model.currentTask = 0
	model.states[0] = TaskRunning

	msg := taskDoneMsg{index: 0, err: WarnErr{Msg: "warning"}}
	newModel, _ := model.Update(msg)
	m := newModel.(MultiTaskModel)

	if m.states[0] != TaskWarning {
		t.Errorf("states[0] = %v, want TaskWarning", m.states[0])
	}

	if m.done {
		t.Error("done should be false after warning (should continue)")
	}
}

func TestMultiTaskModel_Update_KeyMsg(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test", tasks)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("ctrl+c should return tea.Quit command")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd = model.Update(msg)

	if cmd == nil {
		t.Error("q should return tea.Quit command")
	}
}

func TestMultiTaskModel_View(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return nil }},
		{Name: "Task 2", Run: func() error { return nil }},
	}

	model := NewMultiTaskModel("Test Title", tasks)
	model.states[0] = TaskSuccess
	model.states[1] = TaskPending

	view := model.View()

	if !strings.Contains(view, "Test Title") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "Task 1") {
		t.Error("View should contain Task 1")
	}

	if !strings.Contains(view, "Task 2") {
		t.Error("View should contain Task 2")
	}

	if !strings.Contains(view, SymbolSuccess) {
		t.Error("View should contain success symbol")
	}
}

func TestMultiTaskModel_HasError(t *testing.T) {
	tasks := []Task{
		{Name: "Task 1", Run: func() error { return nil }},
		{Name: "Task 2", Run: func() error { return errors.New("error") }},
	}

	model := NewMultiTaskModel("Test", tasks)

	if model.HasError() {
		t.Error("HasError() should be false initially")
	}

	model.states[0] = TaskWarning
	model.errors[0] = WarnErr{Msg: "warning"}

	if model.HasError() {
		t.Error("HasError() should be false for warnings only")
	}

	model.states[1] = TaskFailed
	model.errors[1] = errors.New("error")

	if !model.HasError() {
		t.Error("HasError() should be true after failure")
	}
}

func TestSingleTaskModel_Init(t *testing.T) {
	task := Task{Name: "Test", Run: func() error { return nil }}
	model := NewSingleTaskModel("Testing...", task)

	if model.state != TaskPending {
		t.Errorf("state = %v, want TaskPending", model.state)
	}

	if model.message != "Testing..." {
		t.Errorf("message = %q, want %q", model.message, "Testing...")
	}
}

func TestSingleTaskModel_Update_TaskDone(t *testing.T) {
	task := Task{Name: "Test", Run: func() error { return nil }}
	model := NewSingleTaskModel("Testing...", task)
	model.state = TaskRunning

	msg := taskDoneMsg{index: 0, err: nil}
	newModel, _ := model.Update(msg)
	m := newModel.(SingleTaskModel)

	if m.state != TaskSuccess {
		t.Errorf("state = %v, want TaskSuccess", m.state)
	}
}

func TestSingleTaskModel_Update_TaskFailed(t *testing.T) {
	task := Task{Name: "Test", Run: func() error { return errors.New("failed") }}
	model := NewSingleTaskModel("Testing...", task)
	model.state = TaskRunning

	msg := taskDoneMsg{index: 0, err: errors.New("failed")}
	newModel, _ := model.Update(msg)
	m := newModel.(SingleTaskModel)

	if m.state != TaskFailed {
		t.Errorf("state = %v, want TaskFailed", m.state)
	}

	if m.err == nil || m.err.Error() != "failed" {
		t.Errorf("err = %v, want 'failed'", m.err)
	}
}

func TestSingleTaskModel_View(t *testing.T) {
	task := Task{Name: "Test", Run: func() error { return nil }}
	model := NewSingleTaskModel("Testing...", task)

	view := model.View()

	if !strings.Contains(view, "Testing...") {
		t.Error("View should contain message")
	}

	model.state = TaskSuccess
	view = model.View()

	if !strings.Contains(view, SymbolSuccess) {
		t.Error("View should contain success symbol")
	}

	model.state = TaskFailed
	model.err = errors.New("test error")
	view = model.View()

	if !strings.Contains(view, "test error") {
		t.Error("View should contain error message")
	}
}
