package commit

import (
	"bufio"
	"fmt"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"os"
	"regexp"
	"strings"
)

func CheckCommitMessage(args []string) {
	if len(args) != 1 {
		checkFailed()
	}

	reg := regexp.MustCompile(consts.CommitMessagePattern)

	f, err := os.Open(args[0])
	defer f.Close()
	util.CheckAndExit(err)

	r := bufio.NewReader(f)
	firstLine, err := r.ReadString('\n')
	util.CheckAndExit(err)
	if firstLine == "" {
		checkFailed()
	} else {
		firstLine = strings.Replace(firstLine, "\n", "", -1)
	}
	commitTypes := reg.FindAllStringSubmatch(firstLine, -1)
	if len(commitTypes) != 1 {
		checkFailed()
	} else {
		switch commitTypes[0][1] {
		case string(consts.FEAT):
			fallthrough
		case string(consts.FIX):
			fallthrough
		case string(consts.DOCS):
			fallthrough
		case string(consts.STYLE):
			fallthrough
		case string(consts.REFACTOR):
			fallthrough
		case string(consts.TEST):
			fallthrough
		case string(consts.CHORE):
			fallthrough
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
