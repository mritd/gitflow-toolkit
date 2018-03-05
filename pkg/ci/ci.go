package ci

import (
	"github.com/mritd/gitflow-toolkit/pkg/config"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"gopkg.in/AlecAivazis/survey.v1"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// 检查当前位置是否为 git 项目
func CheckGitProject() bool {
	return exec.Command("git", "rev-parse", "--show-toplevel").Run() == nil
}

// 选择提交类型
func SelectCommitType() config.CommitMessage {
	cm := config.CommitMessage{}
	prompt := &survey.Select{
		Message: "请选择提交类型:",
		Options: []string{
			config.COMMIT_TYPE_MSG_FEAT,
			config.COMMIT_TYPE_MSG_FIX,
			config.COMMIT_TYPE_MSG_DOCS,
			config.COMMIT_TYPE_MSG_STYLE,
			config.COMMIT_TYPE_MSG_REFACTOR,
			config.COMMIT_TYPE_MSG_TEST,
			config.COMMIT_TYPE_MSG_CHORE,
			config.COMMIT_TYPE_MSG_PERF,
			config.COMMIT_TYPE_MSG_EXIT,
		},
		Default:  config.COMMIT_TYPE_MSG_FEAT,
		PageSize: 9,
	}
	err := survey.AskOne(prompt, &cm, nil)
	util.CheckAndExit(err)
	return cm
}

// 输入影响范围
func InputScope() string {
	scope := ""
	prompt := &survey.Input{
		Message: "请输入本次提交影响范围:",
	}
	err := survey.AskOne(prompt, &scope, nil)
	util.CheckAndExit(err)
	return scope

}

// 输入提交主题
func InputSubject() string {
	subject := ""
	prompt := &survey.Input{
		Message: "请输入本次提交简短描述(不能超过 50 个字):",
	}
	err := survey.AskOne(prompt, &subject, nil)
	util.CheckAndExit(err)
	return subject
}

// 输入完整提交信息
func InputBody() string {
	body := ""
	prompt := &survey.Input{
		Message: "请输入本次提交完整的提交信息:",
	}
	err := survey.AskOne(prompt, &body, nil)
	util.CheckAndExit(err)
	if body == "edit" {
		body = InputBigBody()
	}
	return body
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
	footer := ""
	prompt := &survey.Input{
		Message: "输入本次提交可以解决/关闭的相关问题，建议使用关键字 refs、close:",
	}
	err := survey.AskOne(prompt, &footer, nil)
	util.CheckAndExit(err)
	return footer
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

func Commit(cm *config.CommitMessage) {
	t, err := template.New("commitMessage").Parse(config.CommitTpl)
	util.CheckAndExit(err)
	t.Execute(os.Stdout, cm)
}
