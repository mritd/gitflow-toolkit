package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/utils"
)

func CheckCommitMessage(file string) {

	reg := regexp.MustCompile(consts.CommitMessagePattern)

	b, err := ioutil.ReadFile(file)
	utils.CheckAndExit(err)

	commitTypes := reg.FindAllStringSubmatch(string(b), -1)
	if len(commitTypes) != 1 {
		checkFailed()
	} else {
		switch commitTypes[0][1] {
		case string(consts.FEAT):
		case string(consts.FIX):
		case string(consts.DOCS):
		case string(consts.STYLE):
		case string(consts.REFACTOR):
		case string(consts.TEST):
		case string(consts.CHORE):
		case string(consts.PERF):
		case string(consts.HOTFIX):
		default:
			if !strings.HasPrefix(string(b), "Merge branch") {
				checkFailed()
			}
		}
	}
}

func checkFailed() {
	fmt.Println("check commit message style failed")
	os.Exit(1)
}
