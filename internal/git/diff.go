package git

import (
	"fmt"
	"strings"
)

// FileDiff represents a single file's diff.
type FileDiff struct {
	Path string
	Diff string
}

// GetStagedDiff returns the staged diff with specified context lines.
func GetStagedDiff(context int) (string, error) {
	return Run("diff", "--staged", fmt.Sprintf("-U%d", context))
}

// GetPreviousCommits returns the last n commit messages.
// If on a feature branch, returns commits since diverging from main.
func GetPreviousCommits(n int) ([]string, error) {
	// Get main branch name
	mainBranch := getMainBranch()

	// Get current branch
	currentBranch, err := CurrentBranch()
	if err != nil {
		currentBranch = ""
	}

	var output string
	if currentBranch != "" && currentBranch != mainBranch {
		// On feature branch, get commits since diverging from main
		output, err = Run("log", fmt.Sprintf("%s..", mainBranch), "--pretty=format:%s", fmt.Sprintf("-n%d", n))
		if err != nil || output == "" {
			// Fallback to recent commits
			output, _ = Run("log", "--pretty=format:%s", fmt.Sprintf("-n%d", n))
		}
	} else {
		// On main branch or can't determine, get recent commits
		output, _ = Run("log", "--pretty=format:%s", fmt.Sprintf("-n%d", n))
	}

	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	var commits []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			commits = append(commits, line)
		}
	}
	return commits, nil
}

// getMainBranch detects the main branch name.
func getMainBranch() string {
	// Try to get from remote HEAD
	if ref, err := Run("symbolic-ref", "refs/remotes/origin/HEAD"); err == nil {
		ref = strings.TrimPrefix(ref, "refs/remotes/origin/")
		if ref != "" {
			return ref
		}
	}

	// Fallback: check if 'main' exists
	if _, err := Run("rev-parse", "--verify", "main"); err == nil {
		return "main"
	}

	// Fallback: check if 'master' exists
	if _, err := Run("rev-parse", "--verify", "master"); err == nil {
		return "master"
	}

	return "main"
}

// SplitDiffByFile splits a unified diff into per-file chunks.
func SplitDiffByFile(diff string) []FileDiff {
	if diff == "" {
		return nil
	}

	var result []FileDiff
	var currentPath string
	var currentDiff strings.Builder

	lines := strings.Split(diff, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			// Save previous chunk
			if currentPath != "" {
				result = append(result, FileDiff{
					Path: currentPath,
					Diff: strings.TrimSuffix(currentDiff.String(), "\n"),
				})
			}

			// Extract file path from "diff --git a/path b/path"
			currentPath = extractFilePath(line)
			currentDiff.Reset()
			currentDiff.WriteString(line)
			currentDiff.WriteString("\n")
		} else if currentPath != "" {
			currentDiff.WriteString(line)
			// Only add newline if not the last line
			if i < len(lines)-1 {
				currentDiff.WriteString("\n")
			}
		}
	}

	// Save last chunk
	if currentPath != "" {
		result = append(result, FileDiff{
			Path: currentPath,
			Diff: strings.TrimSuffix(currentDiff.String(), "\n"),
		})
	}

	return result
}

// extractFilePath extracts the file path from a diff header line.
// Input: "diff --git a/path/to/file b/path/to/file"
// Output: "path/to/file"
func extractFilePath(line string) string {
	// Remove "diff --git " prefix
	line = strings.TrimPrefix(line, "diff --git ")

	// Split by " b/"
	parts := strings.Split(line, " b/")
	if len(parts) >= 2 {
		return parts[1]
	}

	// Fallback: try to extract from "a/path"
	if strings.HasPrefix(line, "a/") {
		parts = strings.SplitN(line, " ", 2)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "a/")
		}
	}

	return line
}
