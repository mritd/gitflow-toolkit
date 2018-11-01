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
	"encoding/base64"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var bannerBase64 = "IC44ODg4OC4gIG9vICAgZFAgICAgICAgODg4ODg4ODhiIGRQICAgICAgICAgICAgICAgICAgICAgCmQ4JyAgIGA4OCAgICAgIDg4ICAgICAgIDg4ICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgIAo4OCAgICAgICAgZFAgZDg4ODhQICAgIGE4OGFhYWEgICAgODggLmQ4ODg4Yi4gZFAgIGRQICBkUCAKODggICBZUDg4IDg4ICAgODggICAgICAgODggICAgICAgIDg4IDg4JyAgYDg4IDg4ICA4OCAgODggClk4LiAgIC44OCA4OCAgIDg4ICAgICAgIDg4ICAgICAgICA4OCA4OC4gIC44OCA4OC44OGIuODgnIAogYDg4ODg4JyAgZFAgICBkUCAgICAgICBkUCAgICAgICAgZFAgYDg4ODg4UCcgODg4OFAgWThQICA="

var versionTpl = `
%s

Name: gitflow-toolkit
Version: %s
Arch: %s
BuildTime: %s
CommitID: %s
`

var (
	Version   string
	BuildTime string
	CommitID  string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long: `
Print version.`,
	Run: func(cmd *cobra.Command, args []string) {
		banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
		fmt.Printf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildTime, CommitID)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
