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

func commandFor(basename string, rootCommand *cobra.Command) *cobra.Command {

	c, _, _ := rootCommand.Find([]string{basename})
	if c != nil {
		// 从顶级 cmd 移除子 cmd，以保证子 cmd HasParent()=false
		rootCommand.RemoveCommand(c)
		return c
	}
	return rootCommand
}

func main() {

	basename := filepath.Base(os.Args[0])
	util.CheckAndExit(commandFor(basename, cmd.RootCmd).Execute())
}
