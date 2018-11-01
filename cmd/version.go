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

var bannerBase64 = "CiAgLm9vb29vby4gICAgIG84byAgICAgIC4gICAgICBvb29vb29vb29vb28gb29vbwogZDhQJyAgYFk4YiAgICBgIicgICAgLm84ICAgICAgYDg4OCcgICAgIGA4IGA4ODgKODg4ICAgICAgICAgICBvb29vICAubzg4OG9vICAgICA4ODggICAgICAgICAgODg4ICAgLm9vb29vLiAgb29vbyBvb29vICAgIG9vbwo4ODggICAgICAgICAgIGA4ODggICAgODg4ICAgICAgIDg4OG9vb284ICAgICA4ODggIGQ4OCcgYDg4YiAgYDg4LiBgODguICAuOCcKODg4ICAgICBvb29vbyAgODg4ICAgIDg4OCAgICAgICA4ODggICAgIiAgICAgODg4ICA4ODggICA4ODggICBgODguLl04OC4uOCcKYDg4LiAgICAuODgnICAgODg4ICAgIDg4OCAuICAgICA4ODggICAgICAgICAgODg4ICA4ODggICA4ODggICAgYDg4OCdgODg4JwogYFk4Ym9vZDhQJyAgIG84ODhvICAgIjg4OCIgICAgbzg4OG8gICAgICAgIG84ODhvIGBZOGJvZDhQJyAgICAgYDgnICBgOCc="

var versionTpl = `%s

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
