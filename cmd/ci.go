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
		cm := &ci.CommitMessage{}
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
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ciCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ciCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
