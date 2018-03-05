package config

type CommitType string

const (
	COMMIT_TYPE_MSG_FEAT     = "1、feat: 新功能 (Introducing new features)"
	COMMIT_TYPE_MSG_FIX      = "2、fix: 修补bug (Fixing a bug)"
	COMMIT_TYPE_MSG_DOCS     = "3、docs: 文档 (Writing docs)"
	COMMIT_TYPE_MSG_STYLE    = "4、style: 格式 (Improving structure/Format of the code)"
	COMMIT_TYPE_MSG_REFACTOR = "5、refactor: 重构 (Refactoring code)"
	COMMIT_TYPE_MSG_TEST     = "6、test: 增加测试 (Adding tests)"
	COMMIT_TYPE_MSG_CHORE    = "7、chore: 构建过程或辅助工具的变动 (Changing configuration files)"
	COMMIT_TYPE_MSG_PERF     = "8、perf:  性能优化 (Improving performance)"
	COMMIT_TYPE_MSG_EXIT     = "9、exit: 退出本次提交 (Exit)"
)

const (
	COMMIT_TYPE_FEAT     CommitType = "feat"
	COMMIT_TYPE_FIX      CommitType = "fix"
	COMMIT_TYPE_DOCS     CommitType = "docs"
	COMMIT_TYPE_STYLE    CommitType = "style"
	COMMIT_TYPE_REFACTOR CommitType = "refactor"
	COMMIT_TYPE_TEST     CommitType = "test"
	COMMIT_TYPE_CHORE    CommitType = "chore"
	COMMIT_TYPE_PERF     CommitType = "perf"
	COMMIT_TYPE_EXIT     CommitType = "exit"
)

const CommitTpl = `{{ .Type }}({{ .Scope }}): {{ .Subject }}

{{ .Body }}

{{ .Footer }}

{{ .Sob }}
`

type CommitMessage struct {
	Type    CommitType
	Scope   string
	Subject string
	Body    string
	Footer  string
	Sob     string
}

func (cm *CommitMessage) WriteAnswer(name string, value interface{}) error {
	ct := value.(string)

	switch ct {
	case COMMIT_TYPE_MSG_FEAT:
		cm.Type = COMMIT_TYPE_FEAT
	case COMMIT_TYPE_MSG_FIX:
		cm.Type = COMMIT_TYPE_FIX
	case COMMIT_TYPE_MSG_DOCS:
		cm.Type = COMMIT_TYPE_DOCS
	case COMMIT_TYPE_MSG_STYLE:
		cm.Type = COMMIT_TYPE_STYLE
	case COMMIT_TYPE_MSG_REFACTOR:
		cm.Type = COMMIT_TYPE_REFACTOR
	case COMMIT_TYPE_MSG_TEST:
		cm.Type = COMMIT_TYPE_TEST
	case COMMIT_TYPE_MSG_CHORE:
		cm.Type = COMMIT_TYPE_CHORE
	case COMMIT_TYPE_MSG_PERF:
		cm.Type = COMMIT_TYPE_PERF
	case COMMIT_TYPE_MSG_EXIT:
		cm.Type = COMMIT_TYPE_EXIT
	}
	return nil
}
