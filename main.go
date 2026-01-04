package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/mritd/gitflow-toolkit/v3/cmd"
	"github.com/mritd/gitflow-toolkit/v3/consts"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	// Detect invocation name for git subcommand support
	binName := detectBinaryName()

	// Handle git subcommand invocations (e.g., git-ci -> ci)
	if strings.HasPrefix(binName, consts.GitCommandPrefix) {
		subCmd := strings.TrimPrefix(binName, consts.GitCommandPrefix)
		os.Args = append([]string{os.Args[0], subCmd}, os.Args[1:]...)
	}

	// Set version info
	cmd.SetVersionInfo(cmd.VersionInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	})

	// Execute
	cmd.Execute()
}

// detectBinaryName returns the base name of the executed binary.
func detectBinaryName() string {
	bin, err := exec.LookPath(os.Args[0])
	if err != nil {
		return filepath.Base(os.Args[0])
	}
	return filepath.Base(bin)
}

// See: https://github.com/charmbracelet/lipgloss/issues/40#issuecomment-891167509
func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}
