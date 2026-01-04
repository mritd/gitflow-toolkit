// Package cmd contains all CLI commands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// VersionInfo holds version information for display.
type VersionInfo struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}

var versionInfo = VersionInfo{
	Version: "dev",
}

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
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// renderError renders an error using the Result component and silences cobra's default output.
func renderError(cmd *cobra.Command, title string, err error) error {
	r := common.Error(title, err.Error())
	fmt.Print(common.RenderResult(r))
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return err
}

func init() {
	// Disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// SetVersionInfo sets the version information.
func SetVersionInfo(info VersionInfo) {
	versionInfo = info
	rootCmd.Version = info.Version
	// Set custom version template after version info is populated
	rootCmd.SetVersionTemplate(renderVersion())
}

// renderVersion renders a styled version output.
func renderVersion() string {
	var sb strings.Builder

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorTitleBg).
		Bold(true).
		Padding(0, 1)

	// Label style (width includes space after colon)
	labelStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		Width(14)

	// Value style
	valueStyle := lipgloss.NewStyle().
		Foreground(common.ColorSuccess)

	// Commit style (dimmer for long hash)
	commitStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted)

	// Layout
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 0, 2)
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)

	sb.WriteString(titleLayout.Render(titleStyle.Render("gitflow-toolkit")))
	sb.WriteString("\n\n")

	// Version info rows
	sb.WriteString(contentLayout.Render(labelStyle.Render("Version:") + valueStyle.Render(versionInfo.Version)))
	sb.WriteString("\n")
	sb.WriteString(contentLayout.Render(labelStyle.Render("Built:") + valueStyle.Render(versionInfo.BuildDate)))
	sb.WriteString("\n")
	sb.WriteString(contentLayout.Render(labelStyle.Render("Go Version:") + valueStyle.Render(versionInfo.GoVersion)))
	sb.WriteString("\n")
	sb.WriteString(contentLayout.Render(labelStyle.Render("Commit Hash:") + commitStyle.Render(versionInfo.Commit)))
	sb.WriteString("\n\n")

	return sb.String()
}
