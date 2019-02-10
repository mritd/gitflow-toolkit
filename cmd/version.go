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

// Print toolkit version
func NewVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long: `
Print version.`,
		Run: func(cmd *cobra.Command, args []string) {
			banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
			fmt.Printf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildTime, CommitID)
		},
	}
}
