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
	DefaultAskTpl         = "{{ . | cyan }} "
	DefaultPromptTpl      = "{{ . | green }} "
	DefaultInvalidTpl     = "{{ . | red }} "
	DefaultValidTpl       = "{{ . | green }} "
	DefaultErrorMsgTpl    = "{{ . | red }} "
)

type Prompt struct {
	Config
	Ask     string
	Prompt  string
	FuncMap template.FuncMap

	isFirstRun bool

	ask      *template.Template
	prompt   *template.Template
	valid    *template.Template
	invalid  *template.Template
	errorMsg *template.Template
}

type Config struct {
	AskTpl        string
	PromptTpl     string
	ValidTpl      string
	InvalidTpl    string
	ErrorMsgTpl   string
	CheckListener func(line []rune) error
}

func NewDefaultConfig(check func(line []rune) error) Config {
	return Config{
		AskTpl:        DefaultAskTpl,
		PromptTpl:     DefaultPromptTpl,
		InvalidTpl:    DefaultInvalidTpl,
		ValidTpl:      DefaultValidTpl,
		ErrorMsgTpl:   DefaultErrorMsgTpl,
		CheckListener: check,
	}
}

func NewDefaultPrompt(check func(line []rune) error, ask string) Prompt {
	return Prompt{
		Ask:     ask,
		Prompt:  DefaultPrompt,
		FuncMap: FuncMap,
		Config:  NewDefaultConfig(check),
	}
}

func (p *Prompt) prepareTemplates() {

	var err error
	p.ask, err = template.New("").Funcs(FuncMap).Parse(p.AskTpl)
	util.CheckAndExit(err)
	p.prompt, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl)
	util.CheckAndExit(err)
	p.valid, err = template.New("").Funcs(FuncMap).Parse(p.ValidTpl)
	util.CheckAndExit(err)
	p.invalid, err = template.New("").Funcs(FuncMap).Parse(p.InvalidTpl)
	util.CheckAndExit(err)
	p.errorMsg, err = template.New("").Funcs(FuncMap).Parse(p.ErrorMsgTpl)
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
	// clear line
	case readline.CharInterrupt:
		fmt.Print(moveDown)
		fmt.Print(clearLine)
		fmt.Print(moveUp)
		return r, true
	}
	return r, true
}

func (p *Prompt) Run() string {
	p.isFirstRun = true
	p.prepareTemplates()

	displayPrompt := append(render(p.prompt, p.Prompt), render(p.ask, p.Ask)...)
	validPrompt := append(render(p.valid, p.Prompt), render(p.ask, p.Ask)...)
	invalidPrompt := append(render(p.invalid, p.Prompt), render(p.ask, p.Ask)...)

	l, err := readline.NewEx(&readline.Config{
		Prompt:                 string(displayPrompt),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		FuncFilterInputRune:    filterInput,
	})
	util.CheckAndExit(err)

	l.Config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		// Real-time verification
		if err = p.CheckListener(line); err != nil {
			l.SetPrompt(string(invalidPrompt))
			l.Refresh()
		} else {
			l.SetPrompt(string(validPrompt))
			l.Refresh()
		}
		return nil, 0, false
	})
	defer l.Close()

	// read line
	for {
		if !p.isFirstRun {
			fmt.Print(move2Up)
		}
		s, err := l.Readline()
		util.CheckAndExit(err)
		if err = p.CheckListener([]rune(s)); err != nil {
			fmt.Print(clearLine)
			fmt.Println(string(render(p.errorMsg, DefaultErrorMsgPrefix+err.Error())))
			p.isFirstRun = false
		} else {
			return s
		}
	}
}
