package util

import (
	"strings"

	"github.com/tsuyoshiwada/go-gitcmd"
)

func Git() gitcmd.Client {
	return gitcmd.New(nil)
}

func CheckGitProject() bool {
	_, err := Git().Exec("rev-parse", "--show-toplevel")
	return err == nil
}

func CheckStagedFiles() bool {
	output, _ := Git().Exec("diff", "--cached", "--name-only")
	return strings.Replace(output, " ", "", -1) != ""
}
