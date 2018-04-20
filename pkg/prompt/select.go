package prompt

import (
	"bytes"
	"text/template"

	"github.com/mritd/idgen/util"
)

const (
	DefaultActiveTpl       = "{{ . | cyan }}"
	DefaultInactiveTpl     = "{{ . | white }}"
	DefaultDetailsTpl      = "{{ . | white }}"
	DefaultSelectHeaderTpl = `{{ "Use the arrow keys to navigate: ↓ ↑ → ←" | faint }}`
	DefaultSelectPromptTpl = `{{ "Select" . | faint }}:`
	DefaultSelectAsk       = "Please select"
	NextLine               = "\n"
)

type Select struct {
	SelectConfig
	Items        interface{}
	SelectPrompt string

	selectPrompt *template.Template
	selectHeader *template.Template
	active       *template.Template
	inactive     *template.Template
	details      *template.Template
}

type SelectConfig struct {
	Prompt
	SelectHeaderTpl string
	SelectPromptTpl string
	ActiveTpl       string
	InactiveTpl     string
	DetailsTpl      string
}

func NewDefaultSelectConfig(check func(line []rune) error, selectAsk string) SelectConfig {
	return SelectConfig{
		Prompt:          NewDefaultPrompt(check, selectAsk),
		SelectHeaderTpl: DefaultSelectHeaderTpl,
		SelectPromptTpl: DefaultSelectPromptTpl,
		ActiveTpl:       DefaultActiveTpl,
		InactiveTpl:     DefaultInactiveTpl,
		DetailsTpl:      DefaultDetailsTpl,
	}
}

func NewDefaultSelect(check func(line []rune) error, items interface{}) *Select {
	return &Select{
		SelectConfig: NewDefaultSelectConfig(check, DefaultSelectAsk),
		Items:        items,
	}
}

func (s *Select) prepareTemplates() {

	var err error

	// Prompt prepare
	s.ask, err = template.New("").Funcs(FuncMap).Parse(s.AskTpl)
	util.CheckAndExit(err)
	s.prompt, err = template.New("").Funcs(FuncMap).Parse(s.PromptTpl)
	util.CheckAndExit(err)
	s.valid, err = template.New("").Funcs(FuncMap).Parse(s.ValidTpl)
	util.CheckAndExit(err)
	s.invalid, err = template.New("").Funcs(FuncMap).Parse(s.InvalidTpl)
	util.CheckAndExit(err)
	s.errorMsg, err = template.New("").Funcs(FuncMap).Parse(s.ErrorMsgTpl)
	util.CheckAndExit(err)

	// Select prepare
	s.selectHeader, err = template.New("").Funcs(FuncMap).Parse(s.SelectHeaderTpl)
	util.CheckAndExit(err)
	s.selectPrompt, err = template.New("").Funcs(FuncMap).Parse(s.SelectPromptTpl)
	util.CheckAndExit(err)
	s.active, err = template.New("").Funcs(FuncMap).Parse(s.ActiveTpl)
	util.CheckAndExit(err)
	s.inactive, err = template.New("").Funcs(FuncMap).Parse(s.InactiveTpl)
	util.CheckAndExit(err)
	s.details, err = template.New("").Funcs(FuncMap).Parse(s.DetailsTpl)
	util.CheckAndExit(err)

}

func (s *Select) prepareDisplayData(sindex int) {
	var data bytes.Buffer

	// select header
	data.Write(render(s.selectHeader, NextLine))
	// select prompt
	data.Write(render(s.selectPrompt, s.SelectPrompt+NextLine))
	//for it := range s.Items {
	//
	//}
}
