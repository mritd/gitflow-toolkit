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

package main

import (
	"os"
	"path/filepath"

	"github.com/mritd/gitflow-toolkit/cmd"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/spf13/cobra"
)

func NewGitFlowToolKitCommands() (*cobra.Command, []func() *cobra.Command) {

	allCommandFns := []func() *cobra.Command{
		cmd.NewCi,
		cmd.NewCm,
		cmd.NewInstall,
		cmd.NewUninstall,
	}

	rootCmd := cmd.NewGitFlowToolKit()

	for i := range allCommandFns {
		rootCmd.AddCommand(allCommandFns[i]())
	}

	return rootCmd, allCommandFns
}

// 由于 cobra 在进行执行时永远会在顶级 command 进行 Exec
// 所以在使用文件名进行 case 子 command 时
// 必须保证单独运行的子 command 不具备 Parent
// 这也是这里采用比较 low 的查找方式的方法(后续改进)
func commandFor(basename string, defaultCommand *cobra.Command, commandFns []func() *cobra.Command) *cobra.Command {
	for _, commandFn := range commandFns {
		command := commandFn()
		if command.Name() == basename {
			return command
		}
		for _, alias := range command.Aliases {
			if alias == basename {
				return command
			}
		}
	}
	return defaultCommand
}

func main() {

	rootCmd, allCommandFns := NewGitFlowToolKitCommands()
	basename := filepath.Base(os.Args[0])
	util.CheckAndExit(commandFor(basename, rootCmd, allCommandFns).Execute())
}
