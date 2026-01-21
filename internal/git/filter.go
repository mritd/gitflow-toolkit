package git

import (
	"path/filepath"
	"strings"
)

// lockFiles contains common package manager lock file names.
var lockFiles = map[string]bool{
	// JavaScript/Node
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"bun.lockb":         true,
	// Go
	"go.sum": true,
	// Rust
	"Cargo.lock": true,
	// Ruby
	"Gemfile.lock": true,
	// Python
	"poetry.lock":  true,
	"Pipfile.lock": true,
	"uv.lock":      true,
	"pdm.lock":     true,
	// PHP
	"composer.lock": true,
	// Dart/Flutter
	"pubspec.lock": true,
	// .NET
	"packages.lock.json": true,
}

// IsLockFile checks if the given path is a package manager lock file.
func IsLockFile(path string) bool {
	return lockFiles[filepath.Base(path)]
}

// highPriorityExts contains source code file extensions.
var highPriorityExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true,
	".jsx": true, ".tsx": true, ".java": true, ".rs": true,
	".c": true, ".cpp": true, ".h": true, ".hpp": true,
	".swift": true, ".kt": true, ".scala": true, ".rb": true,
	".ex": true, ".exs": true, ".cs": true, ".php": true,
	".vue": true, ".svelte": true,
}

// mediumPriorityExts contains config/script file extensions.
var mediumPriorityExts = map[string]bool{
	".yaml": true, ".yml": true, ".json": true, ".toml": true,
	".xml": true, ".sh": true, ".bash": true, ".zsh": true,
	".ini": true, ".cfg": true, ".conf": true,
}

// mediumPriorityNames contains config file names (case-insensitive base match).
var mediumPriorityNames = map[string]bool{
	"makefile":       true,
	"cmakelists.txt": true,
}

// mediumPriorityPrefixes contains config file name prefixes.
var mediumPriorityPrefixes = []string{
	"taskfile",
	"dockerfile",
}

// lowPriorityExts contains test/doc file extensions.
var lowPriorityExts = map[string]bool{
	".md": true, ".txt": true, ".rst": true,
	".adoc": true, ".html": true, ".css": true,
}

// lowPriorityPaths contains directory patterns for low priority.
var lowPriorityPaths = []string{
	"docs/", "doc/", "test/", "tests/", "spec/", "specs/",
	"vendor/", "node_modules/", "third_party/", "__pycache__/",
	"dist/", "build/", "target/",
}

// DetectPriority determines the priority level of a file based on its path.
func DetectPriority(path string) FilePriority {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))
	lowerBase := strings.ToLower(base)
	lowerPath := strings.ToLower(path)

	// Check low priority paths first (directories)
	for _, p := range lowPriorityPaths {
		if strings.Contains(lowerPath, p) {
			return PriorityLow
		}
	}

	// Check test file patterns
	if strings.HasSuffix(lowerBase, "_test.go") ||
		strings.Contains(lowerBase, ".spec.") ||
		strings.Contains(lowerBase, ".test.") {
		return PriorityLow
	}

	// Check medium priority by name (before extension check to handle CMakeLists.txt)
	if mediumPriorityNames[lowerBase] {
		return PriorityMedium
	}

	// Check medium priority by prefix
	for _, prefix := range mediumPriorityPrefixes {
		if strings.HasPrefix(lowerBase, prefix) {
			return PriorityMedium
		}
	}

	// Check low priority extensions
	if lowPriorityExts[ext] {
		return PriorityLow
	}

	// Check high priority extensions
	if highPriorityExts[ext] {
		return PriorityHigh
	}

	// Check medium priority extensions
	if mediumPriorityExts[ext] {
		return PriorityMedium
	}

	// Default to medium for unknown files
	return PriorityMedium
}

// CountDiffLines counts added and deleted lines in a diff.
// Only counts lines starting with + or - (excluding +++ and --- headers).
func CountDiffLines(diff string) (add, del int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			add++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			del++
		}
	}
	return add, del
}
