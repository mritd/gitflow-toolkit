package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/mritd/gitflow-toolkit/utils"
)

func CheckCommitMessage(file string) {

	reg := regexp.MustCompile(CommitMessagePattern)

	b, err := ioutil.ReadFile(file)
	utils.CheckAndExit(err)

	commitTypes := reg.FindAllStringSubmatch(string(b), -1)
	if len(commitTypes) != 1 {
		checkFailed()
	} else {
		switch commitTypes[0][1] {
		case string(FEAT):
		case string(FIX):
		case string(DOCS):
		case string(STYLE):
		case string(REFACTOR):
		case string(TEST):
		case string(CHORE):
		case string(PERF):
		case string(HOTFIX):
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
