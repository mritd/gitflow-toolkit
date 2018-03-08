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
	"github.com/mritd/gitflow-toolkit/pkg/ci"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/spf13/cobra"
	"os"
)

var fastCommit = false

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {

		if ! ci.CheckGitProject() {
			fmt.Println("Not a git repository (or any of the parent directories): .git")
			os.Exit(1)
		}

		if ! ci.CheckStagedFiles() {
			fmt.Println("No staged any files")
			os.Exit(1)
		}

		cm := &ci.CommitMessage{}
		if fastCommit {
			cm.Type = consts.FEAT
			cm.Scope = "Undefined"
			cm.Scope = ci.InputSubject()
			cm.Body = cm.Scope
			cm.Sob = ci.GenSOB()
			ci.Commit(cm)
		} else {
			cm.Type = ci.SelectCommitType()

			if cm.Type == consts.EXIT {
				fmt.Println("Talk is cheap Show me the code!")
				os.Exit(0)
			}

			cm.Scope = ci.InputScope()
			cm.Subject = ci.InputSubject()
			cm.Body = ci.InputBody()
			if cm.Body == "big" {
				cm.Body = ci.InputBigBody()
			}
			cm.Footer = ci.InputFooter()
			cm.Sob = ci.GenSOB()
			ci.Commit(cm)
		}
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.Flags().BoolVarP(&fastCommit, "fast", "f", false, "快速提交")
}
