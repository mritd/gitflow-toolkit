package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/mritd/gitflow-toolkit/v3/cmd"
	"github.com/mritd/gitflow-toolkit/v3/internal/config"
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
	if strings.HasPrefix(binName, config.GitCommandPrefix) {
		subCmd := strings.TrimPrefix(binName, config.GitCommandPrefix)
		os.Args = append([]string{os.Args[0], subCmd}, os.Args[1:]...)
	}

	// Set version
	cmd.SetVersion(buildVersionString())

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

// buildVersionString builds the version string.
func buildVersionString() string {
	return version + " (" + commit + ") " + buildDate
}

// See: https://github.com/charmbracelet/lipgloss/issues/40#issuecomment-891167509
func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}
