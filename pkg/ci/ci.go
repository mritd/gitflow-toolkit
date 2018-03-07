package ci

import (
	"errors"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/mritd/promptui"
	"gopkg.in/AlecAivazis/survey.v1"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type CommitTypeMessage struct {
	Type          consts.CommitType
	ZHDescription string
	ENDescription string
}

type CommitMessage struct {
	Type    consts.CommitType
	Scope   string
	Subject string
	Body    string
	Footer  string
	Sob     string
}

// 检查当前位置是否为 git 项目
func CheckGitProject() bool {
	return exec.Command("git", "rev-parse", "--show-toplevel").Run() == nil
}

// 选择提交类型
func SelectCommitType() consts.CommitType {

	commitTypes := []CommitTypeMessage{
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
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "❯ {{ .Type | cyan }} ({{ .ENDescription | cyan }})",
		Inactive: "  {{ .Type | white }} ({{ .ENDescription | white }})",
		Selected: "{{ \"❯ Type\" | green }}: {{ .Type }}",
		Details: `
--------- Commit Type ----------
{{ "Type:" | faint }}	{{ .Type }}
{{ "Description:" | faint }}	{{ .ZHDescription }}({{ .ENDescription }})`,
	}

	searcher := func(input string, index int) bool {
		commitType := commitTypes[index]
		cmType := strings.Replace(strings.ToLower(string(commitType.Type)), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(cmType, input)
	}

	prompt := promptui.Select{
		Label:     "Select Commit Type:",
		Items:     commitTypes,
		Templates: templates,
		Size:      9,
		Searcher:  searcher,
	}
	i, _, err := prompt.Run()
	util.CheckAndExit(err)

	return commitTypes[i].Type
}

// 输入影响范围
func InputScope() string {

	validate := func(input string) error {
		reg := regexp.MustCompile("\\s+")
		if input == "" || reg.ReplaceAllString(input, "") == "" {
			return errors.New("scope is blank")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Scope:",
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	util.CheckAndExit(err)
	return result

}

// 输入提交主题
func InputSubject() string {

	validate := func(input string) error {
		reg := regexp.MustCompile("\\s+")
		if input == "" || reg.ReplaceAllString(input, "") == "" {
			return errors.New("subject is blank")
		}
		if r := []rune(input); len(r) > 50 {
			return errors.New("subject too long")
		}

		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Subject:",
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	util.CheckAndExit(err)
	return result
}

// 输入完整提交信息
func InputBody() string {

	validate := func(input string) error {
		reg := regexp.MustCompile("\\s+")
		if input == "" || reg.ReplaceAllString(input, "") == "" {
			return errors.New("body is blank")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Body:",
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	util.CheckAndExit(err)
	return result
}

// 输入超长文本信息
func InputBigBody() string {
	body := ""
	prompt := &survey.Editor{
		Message: "请输入本次提交完整的提交信息:",
	}
	err := survey.AskOne(prompt, &body, nil)
	util.CheckAndExit(err)
	return body
}

// 输入提交关联信息
func InputFooter() string {

	validate := func(input string) error {
		reg := regexp.MustCompile("\\s+")
		if input == "" || reg.ReplaceAllString(input, "") == "" {
			return errors.New("footer is blank")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Footer:",
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	util.CheckAndExit(err)
	return result
}

// 生成 SOB 签名
func GenSOB() string {
	cmd := exec.Command("git", "var", "GIT_AUTHOR_IDENT")
	buf, err := cmd.CombinedOutput()
	util.CheckAndExit(err)
	if string(buf) == "" {
		return ""
	}

	author := "Undefined"
	email := "Undefined"
	authorInfo := strings.Fields(string(buf))

	if authorInfo[0] != "" {
		author = authorInfo[0]
	}
	if authorInfo[1] != "" {
		email = authorInfo[1]
	}

	return "Signed-off-by: " + author + " " + email
}

func Commit(cm *CommitMessage) {
	t, err := template.New("commitMessage").Parse(consts.CommitTpl)
	util.CheckAndExit(err)
	t.Execute(os.Stdout, cm)
}
