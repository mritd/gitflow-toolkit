// Package cmd contains all CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gitflow-toolkit",
	Short: "A Git Flow commit and branch management toolkit",
	Long: `GitFlow Toolkit is a CLI tool that enforces Git Flow conventions.

It standardizes commit messages following the Angular commit message specification
and provides commands for creating type-prefixed branches.

Commit message format:
  type(scope): subject

  body

  footer

Available commit types:
  feat     - Introducing new features
  fix      - Bug fix
  docs     - Writing docs
  style    - Improving structure/format of the code
  refactor - Refactoring code
  test     - When adding missing tests
  chore    - Changing CI/CD
  perf     - Improving performance
  hotfix   - Bug fix urgently`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// SetVersion sets the version string.
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}
