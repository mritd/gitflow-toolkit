package promptx

import (
	"text/template"

	"os"

	"bytes"

	"fmt"

	"github.com/mritd/promptx/list"
	"github.com/mritd/promptx/util"
	"github.com/mritd/readline"
)

const (
	DefaultActiveTpl       = "{{ . | cyan }}\n"
	DefaultInactiveTpl     = "{{ . | white }}\n"
	DefaultDetailsTpl      = "{{ . | white }}\n"
	DefaultSelectedTpl     = "{{ . | cyan }}\n"
	DefaultSelectHeaderTpl = "{{ \"Use the arrow keys to navigate: ↓ ↑ → ←\" | faint }}"
	DefaultSelectPromptTpl = "{{ \"Select\" | faint }} {{ . | faint}}:"
	DefaultDisPlaySize     = 5
	NewLine                = "\n"
)

type Select struct {
	Config *SelectConfig
	Items  interface{}
	buf    bytes.Buffer
	high   uint

	selectPrompt *template.Template
	selectHeader *template.Template
	selected     *template.Template
	active       *template.Template
	inactive     *template.Template
	details      *template.Template
}

type SelectConfig struct {
	ActiveTpl    string
	InactiveTpl  string
	SelectedTpl  string
	DetailsTpl   string
	DisPlaySize  int
	SelectPrompt string

	selectHeaderTpl string
	selectPromptTpl string
}

func (s *Select) prepareTemplates() {

	var err error

	// set default value
	if s.Config.selectHeaderTpl == "" {
		s.Config.selectHeaderTpl = DefaultSelectHeaderTpl
	}
	if s.Config.selectPromptTpl == "" {
		s.Config.selectPromptTpl = DefaultSelectPromptTpl
	}
	if s.Config.SelectedTpl == "" {
		s.Config.SelectedTpl = DefaultSelectedTpl
	}
	if s.Config.ActiveTpl == "" {
		s.Config.ActiveTpl = DefaultActiveTpl
	}
	if s.Config.InactiveTpl == "" {
		s.Config.InactiveTpl = DefaultInactiveTpl
	}
	if s.Config.DetailsTpl == "" {
		s.Config.DetailsTpl = DefaultDetailsTpl
	}
	if s.Config.DisPlaySize < 1 {
		s.Config.DisPlaySize = DefaultDisPlaySize
	}

	// Select prepare
	s.selectHeader, err = template.New("").Funcs(FuncMap).Parse(s.Config.selectHeaderTpl + NewLine)
	util.CheckAndExit(err)
	s.selectPrompt, err = template.New("").Funcs(FuncMap).Parse(s.Config.selectPromptTpl + NewLine)
	util.CheckAndExit(err)
	s.selected, err = template.New("").Funcs(FuncMap).Parse(s.Config.SelectedTpl)
	util.CheckAndExit(err)
	s.active, err = template.New("").Funcs(FuncMap).Parse(s.Config.ActiveTpl + NewLine)
	util.CheckAndExit(err)
	s.inactive, err = template.New("").Funcs(FuncMap).Parse(s.Config.InactiveTpl + NewLine)
	util.CheckAndExit(err)
	s.details, err = template.New("").Funcs(FuncMap).Parse(s.Config.DetailsTpl + NewLine)
	util.CheckAndExit(err)

}

func (s *Select) writeData(l *list.List) {

	// clean buffer
	s.buf.Reset()

	// clean terminal
	for i := uint(0); i < s.high; i++ {
		s.buf.WriteString(moveUp)
		s.buf.WriteString(clearLine)
	}

	// select header
	s.buf.Write(util.Render(s.selectHeader, ""))

	// select prompt
	s.buf.Write(util.Render(s.selectPrompt, s.Config.SelectPrompt))

	items, idx := l.Items()

	for i, item := range items {
		if i == idx {
			s.buf.Write(util.Render(s.active, item))
		} else {
			s.buf.Write(util.Render(s.inactive, item))
		}
	}
	// detail
	s.buf.Write(util.Render(s.details, items[idx]))

	// hide cursor
	s.buf.WriteString(hideCursor)

	// set high
	//s.high = len(strings.Split(s.buf.String(), "\n")) - 1
	s.high = util.GetTerminalHeight()
}

func (s *Select) Run() int {

	s.prepareTemplates()

	dataList, err := list.New(s.Items, s.Config.DisPlaySize)
	util.CheckAndExit(err)

	l, err := readline.NewEx(&readline.Config{
		Prompt:                 "",
		DisableAutoSaveHistory: true,
		HistoryLimit:           -1,
		InterruptPrompt:        "^C",
		UniqueEditLine:         true,
		DisableBell:            true,
		Stdin:                  readline.NewCancelableStdin(os.Stdin),
	})
	defer l.Close()
	util.CheckAndExit(err)

	filterInput := func(r rune) (rune, bool) {
		switch r {
		case readline.CharInterrupt:
			// show cursor
			l.Write([]byte(showCursor))
			l.Refresh()
			return r, true
		case readline.CharEnter:
			return r, true
		case readline.CharNext:
			dataList.Next()
		case readline.CharPrev:
			dataList.Prev()
		case readline.CharForward:
			dataList.PageDown()
		case readline.CharBackward:
			dataList.PageUp()
		// block other key
		default:
			return r, false
		}
		s.writeData(dataList)
		l.Write(s.buf.Bytes())
		l.Refresh()
		return r, true
	}

	l.Config.FuncFilterInputRune = filterInput

	s.writeData(dataList)
	l.Write(s.buf.Bytes())

	_, err = l.Readline()
	util.CheckAndExit(err)

	items, idx := dataList.Items()
	result := items[idx]

	// clean terminal
	s.buf.Reset()
	for i := uint(0); i < s.high; i++ {
		s.buf.WriteString(moveUp)
		s.buf.WriteString(clearLine)
	}
	l.Write(s.buf.Bytes())

	// show cursor
	l.Write([]byte(showCursor))
	l.Refresh()

	fmt.Println(string(util.Render(s.selected, result)))

	return idx

}
