package main

import (
	"fmt"

	"github.com/mritd/promptx"
)

type CommitType string

const (
	FEAT     CommitType = "feat"
	FIX      CommitType = "fix"
	DOCS     CommitType = "docs"
	STYLE    CommitType = "style"
	REFACTOR CommitType = "refactor"
	TEST     CommitType = "test"
	CHORE    CommitType = "chore"
	PERF     CommitType = "perf"
	EXIT     CommitType = "exit"
)

type TypeMessage struct {
	Type          CommitType
	ZHDescription string
	ENDescription string
}

func main() {
	commitTypes := []TypeMessage{
		{Type: FEAT, ZHDescription: "新功能", ENDescription: "Introducing new features"},
		{Type: FIX, ZHDescription: "修复 Bug", ENDescription: "Bug fix"},
		{Type: DOCS, ZHDescription: "添加文档", ENDescription: "Writing docs"},
		{Type: STYLE, ZHDescription: "调整格式", ENDescription: "Improving structure/format of the code"},
		{Type: REFACTOR, ZHDescription: "重构代码", ENDescription: "Refactoring code"},
		{Type: TEST, ZHDescription: "增加测试", ENDescription: "When adding missing tests"},
		{Type: CHORE, ZHDescription: "CI/CD 变动", ENDescription: "Changing CI/CD"},
		{Type: PERF, ZHDescription: "性能优化", ENDescription: "Improving performance"},
		{Type: EXIT, ZHDescription: "退出", ENDescription: "Exit commit"},
	}

	cfg := &promptx.SelectConfig{
		ActiveTpl:    "»  {{ .Type | cyan }} ({{ .ENDescription | cyan }})",
		InactiveTpl:  "  {{ .Type | white }} ({{ .ENDescription | white }})",
		SelectPrompt: "Commit Type",
		SelectedTpl:  "{{ \"» Type:\" | green }} {{ .Type }}",
		DisPlaySize:  9,
		DetailsTpl: `
--------- Commit Type ----------
{{ "Type:" | faint }}	{{ .Type }}
{{ "Description:" | faint }}	{{ .ZHDescription }}({{ .ENDescription }})`,
	}

	s := &promptx.Select{
		Items:  commitTypes,
		Config: cfg,
	}
	fmt.Println(s.Run())
}
