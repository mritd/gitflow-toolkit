// Copyright © 2018 mritd <mritd1234@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	"github.com/mritd/gitflow-toolkit/pkg/commit"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/spf13/cobra"
)

var fastCommit = false

func NewCi() *cobra.Command {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: "交互式输入 commit message",
		Long: `
交互式输入 git commit message，commit message 格式为

<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>

该格式来源于 Angular 社区提交规范`,
		Aliases: []string{"git-ci"},
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckGitProject()

			if !util.CheckStagedFiles() {
				fmt.Println("No staged any files")
				os.Exit(1)
			}

			cm := &commit.Message{Sob: commit.GenSOB()}

			if fastCommit {
				cm.Type = consts.FEAT
				cm.Scope = "Undefined"
				cm.Subject = commit.InputSubject()
				commit.Commit(cm)
			} else {
				cm.Type = commit.SelectCommitType()
				cm.Scope = commit.InputScope()
				cm.Subject = commit.InputSubject()
				cm.Body = commit.InputBody()
				cm.Footer = commit.InputFooter()
				commit.Commit(cm)
			}
		},
	}

	ciCmd.Flags().BoolVarP(&fastCommit, "fast", "f", false, "快速提交")
	return ciCmd
}
