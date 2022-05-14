package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mattn/go-runewidth"
)

var (
	version     string
	buildDate   string
	buildCommit string
)

func main() {
	app := mainApp

	bin, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	binName := filepath.Base(bin)
	for _, sa := range subApps {
		if binName == sa.Name {
			app = sa
		}
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// See also: https://github.com/charmbracelet/lipgloss/issues/40#issuecomment-891167509
func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}
