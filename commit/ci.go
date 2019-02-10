package commit

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/mritd/promptx"
	"github.com/pkg/errors"
)

type TypeMessage struct {
	Type          consts.CommitType
	ZHDescription string
	ENDescription string
}

type Message struct {
	Type    consts.CommitType
	Scope   string
	Subject string
	Body    string
	Footer  string
	Sob     string
}

// 选择提交类型
func SelectCommitType() consts.CommitType {

	commitTypes := []TypeMessage{
		{Type: consts.FEAT, ZHDescription: "新功能", ENDescription: "Introducing new features"},
		{Type: consts.FIX, ZHDescription: "修复 Bug", ENDescription: "Bug fix"},
		{Type: consts.DOCS, ZHDescription: "添加文档", ENDescription: "Writing docs"},
		{Type: consts.STYLE, ZHDescription: "调整格式", ENDescription: "Improving structure/format of the code"},
		{Type: consts.REFACTOR, ZHDescription: "重构代码", ENDescription: "Refactoring code"},
		{Type: consts.TEST, ZHDescription: "增加测试", ENDescription: "When adding missing tests"},
		{Type: consts.CHORE, ZHDescription: "CI/CD 变动", ENDescription: "Changing CI/CD"},
		{Type: consts.PERF, ZHDescription: "性能优化", ENDescription: "Improving performance"},
		{Type: consts.EXIT, ZHDescription: "退出", ENDescription: "Exit commit"},
	}
	cfg := &promptx.SelectConfig{
		ActiveTpl:    "»  {{ .Type | cyan }} ({{ .ENDescription | cyan }})",
		InactiveTpl:  "  {{ .Type | white }} ({{ .ENDescription | white }})",
		SelectPrompt: "Commit Type",
		SelectedTpl:  "{{ \"» \" | green }}{{\"Type:\" | cyan }} {{ .Type }}",
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

	idx := s.Run()

	if commitTypes[idx].Type == consts.EXIT {
		fmt.Println("Talk is cheap. Show me the code.")
		os.Exit(0)
	}

	return commitTypes[idx].Type
}

// 输入影响范围
func InputScope() string {

	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		} else {
			return nil
		}
	}, "Scope:")

	return strings.TrimSpace(p.Run())

}

// 输入提交主题
func InputSubject() string {

	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		} else if len(line) > 50 {
			return errors.New("Input length must < 25!")
		} else {
			return nil
		}
	}, "Subject:")

	return strings.TrimSpace(p.Run())
}

// 输入完整提交信息
func InputBody() string {

	p := promptx.NewDefaultPrompt(func(line []rune) error {
		return nil
	}, "Body:")

	body := strings.TrimSpace(p.Run())
	if body == "big" {
		return util.OSEditInput()
	}

	return body
}

// 输入提交关联信息
func InputFooter() string {

	p := promptx.NewDefaultPrompt(func(line []rune) error {
		return nil
	}, "Footer:")

	return strings.TrimSpace(p.Run())
}

// 生成 SOB 签名
func GenSOB() string {

	author := "Undefined"
	email := "Undefined"

	output := util.MustExecRtOut("git", "var", "GIT_AUTHOR_IDENT")
	authorInfo := strings.Fields(output)

	if len(authorInfo) > 1 && authorInfo[0] != "" {
		author = authorInfo[0]
	}
	if len(authorInfo) > 2 && authorInfo[1] != "" {
		email = authorInfo[1]
	}

	return "Signed-off-by: " + author + " " + email
}

// 提交
func Commit(cm *Message) {

	if cm.Body == "" {
		cm.Body = cm.Subject
	}

	t, err := template.New("commitMessage").Parse(consts.CommitTpl)
	util.CheckAndExit(err)
	f, err := ioutil.TempFile("", "git-commit")
	defer f.Close()
	defer os.Remove(f.Name())
	util.CheckAndExit(err)
	t.Execute(f, cm)
	util.MustExec("git", "commit", "-F", f.Name())

	fmt.Println("\n✔ Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.")
}
