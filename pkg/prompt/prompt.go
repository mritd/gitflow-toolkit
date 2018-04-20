package prompt

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/mritd/readline"
)

const (
	DefaultPrompt         = "»"
	DefaultErrorMsgPrefix = "✘ "
	DefaultQuestionTpl    = "{{ . | cyan }} "
	DefaultPromptTpl      = "{{ . | green }} "
	DefaultInvalidTpl     = "{{ . | red }} "
	DefaultValidTpl       = "{{ . | green }} "
	DefaultErrorMsgTpl    = "{{ . | red }} "
)

type Prompt struct {
	Question  string
	Prompt    string
	PromptTpl *Tpl
	FuncMap   template.FuncMap

	question *template.Template
	prompt   *template.Template
	valid    *template.Template
	invalid  *template.Template
	errorMsg *template.Template
}

type Tpl struct {
	QuestionTpl   string
	PromptTpl     string
	ValidTpl      string
	InvalidTpl    string
	ErrorMsgTpl   string
	CheckListener func(line []rune) error
}

func NewDefaultTpl(check func(line []rune) error) *Tpl {
	return &Tpl{
		QuestionTpl:   DefaultQuestionTpl,
		PromptTpl:     DefaultPromptTpl,
		InvalidTpl:    DefaultInvalidTpl,
		ValidTpl:      DefaultValidTpl,
		ErrorMsgTpl:   DefaultErrorMsgTpl,
		CheckListener: check,
	}
}

func NewDefaultPrompt(check func(line []rune) error, question string) *Prompt {
	return &Prompt{
		Question:  question,
		Prompt:    DefaultPrompt,
		PromptTpl: NewDefaultTpl(check),
		FuncMap:   FuncMap,
	}
}

func (p *Prompt) prepareTemplates() {

	var err error
	p.question, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl.QuestionTpl)
	util.CheckAndExit(err)
	p.prompt, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl.PromptTpl)
	util.CheckAndExit(err)
	p.valid, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl.ValidTpl)
	util.CheckAndExit(err)
	p.invalid, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl.InvalidTpl)
	util.CheckAndExit(err)
	p.errorMsg, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl.ErrorMsgTpl)
	util.CheckAndExit(err)

}

func render(tpl *template.Template, data interface{}) []byte {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		return []byte(fmt.Sprintf("%v", data))
	}
	return buf.Bytes()
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func (p *Prompt) Run() string {

	p.prepareTemplates()

	displayPrompt := append(render(p.prompt, p.Prompt), render(p.question, p.Question)...)
	validPrompt := append(render(p.valid, p.Prompt), render(p.question, p.Question)...)
	invalidPrompt := append(render(p.invalid, p.Prompt), render(p.question, p.Question)...)

	l, err := readline.NewEx(&readline.Config{
		Prompt:                 string(displayPrompt),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		FuncFilterInputRune:    filterInput,
	})
	util.CheckAndExit(err)

	l.Config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		if err = p.PromptTpl.CheckListener(line); err != nil {
			l.SetPrompt(string(invalidPrompt))
			l.Refresh()

		} else {
			l.SetPrompt(string(validPrompt))
			l.Refresh()

		}
		return nil, 0, false

	})
	defer l.Close()
	for {
		s, err := l.Readline()
		util.CheckAndExit(err)
		if err = p.PromptTpl.CheckListener([]rune(s)); err != nil {
			fmt.Println(string(render(p.errorMsg, DefaultErrorMsgPrefix+err.Error())))
		} else {
			return s
		}
	}
}
