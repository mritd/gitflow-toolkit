package git

import (
	"path/filepath"
	"sort"
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

// FileCategory represents the category for sorting in commit generation.
type FileCategory int

const (
	CategoryCoreCode FileCategory = iota // [CORE] marked code files
	CategoryCode                         // Regular code files
	CategoryTest                         // Test files
	CategoryConfig                       // Config files (go.mod, package.json, etc.)
	CategoryDoc                          // Documentation (*.md)
	CategoryOther                        // Other files (.gitignore, etc.)
)

// IsTestFile checks if the file is a test file.
func IsTestFile(path string) bool {
	lowerPath := strings.ToLower(path)
	base := strings.ToLower(filepath.Base(path))

	// Check filename patterns
	if strings.HasSuffix(base, "_test.go") ||
		strings.Contains(base, ".spec.") ||
		strings.Contains(base, ".test.") {
		return true
	}

	// Check directory patterns (with or without leading slash)
	return strings.Contains(lowerPath, "/test/") ||
		strings.Contains(lowerPath, "/tests/") ||
		strings.HasPrefix(lowerPath, "test/") ||
		strings.HasPrefix(lowerPath, "tests/")
}

// IsDocFile checks if the file is a documentation file.
func IsDocFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".txt" || ext == ".rst" || ext == ".adoc"
}

// IsConfigFile checks if the file is a config file.
func IsConfigFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	configNames := map[string]bool{
		"go.mod": true, "go.sum": true,
		"package.json": true, "package-lock.json": true,
		"tsconfig.json": true, "vite.config.ts": true,
		"makefile": true, "dockerfile": true,
		".goreleaser.yaml": true, ".goreleaser.yml": true,
	}
	if configNames[base] {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml" || ext == ".toml" || ext == ".ini"
}

// IsOtherFile checks if the file is an "other" file (gitignore, etc.).
func IsOtherFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".") && !strings.HasSuffix(base, ".go")
}

// GetFileCategory determines the category of a file for sorting.
func GetFileCategory(f FileDiff, isCore bool) FileCategory {
	if isCore && highPriorityExts[strings.ToLower(filepath.Ext(f.Path))] {
		return CategoryCoreCode
	}
	if IsTestFile(f.Path) {
		return CategoryTest
	}
	if highPriorityExts[strings.ToLower(filepath.Ext(f.Path))] {
		return CategoryCode
	}
	if IsConfigFile(f.Path) {
		return CategoryConfig
	}
	if IsDocFile(f.Path) {
		return CategoryDoc
	}
	return CategoryOther
}

// SortFilesForCommit sorts files by category for commit message generation.
// Order: [CORE] code > code > test > config > doc > other
// Within same category, sort by lines changed (descending).
func SortFilesForCommit(files []FileDiff, coreIndices map[int]bool) []FileDiff {
	type sortItem struct {
		index    int
		file     FileDiff
		category FileCategory
		lines    int
	}

	items := make([]sortItem, len(files))
	for i, f := range files {
		items[i] = sortItem{
			index:    i,
			file:     f,
			category: GetFileCategory(f, coreIndices[i]),
			lines:    f.LinesAdd + f.LinesDel,
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].category != items[j].category {
			return items[i].category < items[j].category
		}
		return items[i].lines > items[j].lines // Larger changes first within category
	})

	result := make([]FileDiff, len(files))
	for i, item := range items {
		result[i] = item.file
	}
	return result
}

// coreFilePriority returns a priority score for core file detection.
// Lower score = higher priority: code (0) > config (1) > doc (2) > other (3)
// Test files are excluded (returns -1).
func coreFilePriority(f FileDiff) int {
	if IsTestFile(f.Path) {
		return -1 // exclude test files
	}
	ext := strings.ToLower(filepath.Ext(f.Path))
	if highPriorityExts[ext] {
		return 0 // code files
	}
	if IsConfigFile(f.Path) {
		return 1 // config files
	}
	if IsDocFile(f.Path) {
		return 2 // doc files
	}
	return 3 // other files
}

// DetectCoreFiles returns indices of core files (top N files by priority and lines changed).
// Priority order: code > config > doc > other. Test files are excluded.
func DetectCoreFiles(files []FileDiff, maxCore int) map[int]bool {
	type candidate struct {
		index    int
		priority int
		lines    int
	}

	var candidates []candidate
	for i, f := range files {
		priority := coreFilePriority(f)
		if priority < 0 {
			continue // skip test files
		}
		candidates = append(candidates, candidate{
			index:    i,
			priority: priority,
			lines:    f.LinesAdd + f.LinesDel,
		})
	}

	// Sort by priority (ascending), then by lines changed (descending)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		return candidates[i].lines > candidates[j].lines
	})

	result := make(map[int]bool)
	for i := 0; i < len(candidates) && i < maxCore; i++ {
		result[candidates[i].index] = true
	}
	return result
}

// FilterAndTruncateDiffs filters lock files and truncates diffs by priority.
// Lock files are removed when total file count >= 5.
// Files are truncated by priority: High > Medium > Low.
// Within same priority, smaller files are preserved first.
func FilterAndTruncateDiffs(files []FileDiff, maxLines int) []FileDiff {
	if len(files) == 0 {
		return files
	}

	// Step 1: Parse metadata for all files
	for i := range files {
		files[i].Priority = DetectPriority(files[i].Path)
		files[i].LinesAdd, files[i].LinesDel = CountDiffLines(files[i].Diff)
	}

	// Step 2: Filter lock files if total >= 5
	if len(files) >= 5 {
		filtered := make([]FileDiff, 0, len(files))
		for _, f := range files {
			if !IsLockFile(f.Path) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}

	if len(files) == 0 {
		return files
	}

	// Step 3: Group by priority
	groups := map[FilePriority][]int{
		PriorityHigh:   {},
		PriorityMedium: {},
		PriorityLow:    {},
	}
	for i := range files {
		groups[files[i].Priority] = append(groups[files[i].Priority], i)
	}

	// Step 4: Sort each group by total lines (ascending - small files first)
	for p := range groups {
		sort.Slice(groups[p], func(i, j int) bool {
			a, b := files[groups[p][i]], files[groups[p][j]]
			return (a.LinesAdd + a.LinesDel) < (b.LinesAdd + b.LinesDel)
		})
	}

	// Step 5: Allocate budget by priority
	budget := maxLines
	priorities := []FilePriority{PriorityHigh, PriorityMedium, PriorityLow}

	for _, p := range priorities {
		for _, idx := range groups[p] {
			fileLines := files[idx].LinesAdd + files[idx].LinesDel
			if fileLines <= budget {
				// File fits within budget
				budget -= fileLines
			} else {
				// File exceeds budget - truncate
				files[idx].Truncated = true
				files[idx].Diff = ""
			}
		}
	}

	return files
}
