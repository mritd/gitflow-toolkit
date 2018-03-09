package commit

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
)

func CheckCommitMessage(args []string) {
	if len(args) != 1 {
		checkFailed()
	}

	reg := regexp.MustCompile(consts.CommitMessagePattern)

	f, err := os.Open(args[0])
	defer f.Close()
	util.CheckAndExit(err)

	b, err := ioutil.ReadAll(f)
	util.CheckAndExit(err)

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
		default:
			checkFailed()
		}
	}
}

func checkFailed() {
	fmt.Println("check commit message style failed")
	os.Exit(1)
}
