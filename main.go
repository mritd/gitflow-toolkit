package main

import (
	"os"
	"path/filepath"

	"github.com/mritd/gitflow-toolkit/cmd"
	"github.com/mritd/gitflow-toolkit/utils"
	"github.com/spf13/cobra"
)

func commandFor(basename string, rootCommand *cobra.Command) *cobra.Command {
	c, _, _ := rootCommand.Find([]string{basename})
	if c != nil {
		rootCommand.RemoveCommand(c)
		return c
	}
	return rootCommand
}

func main() {
	basename := filepath.Base(os.Args[0])
	utils.CheckAndExit(commandFor(basename, cmd.RootCmd).Execute())
}
