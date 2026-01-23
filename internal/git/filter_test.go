package git

import "testing"

func TestIsLockFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"package-lock.json", true},
		{"yarn.lock", true},
		{"pnpm-lock.yaml", true},
		{"bun.lockb", true},
		{"go.sum", true},
		{"Cargo.lock", true},
		{"Gemfile.lock", true},
		{"poetry.lock", true},
		{"Pipfile.lock", true},
		{"uv.lock", true},
		{"pdm.lock", true},
		{"composer.lock", true},
		{"pubspec.lock", true},
		{"packages.lock.json", true},
		{"main.go", false},
		{"config.yaml", false},
		{"lock.go", false},
		{"my-lock-file.txt", false},
		{"src/package-lock.json", true},
		{"nested/path/yarn.lock", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsLockFile(tt.path); got != tt.want {
				t.Errorf("IsLockFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDetectPriority(t *testing.T) {
	tests := []struct {
		path string
		want FilePriority
	}{
		// High priority - source code
		{"main.go", PriorityHigh},
		{"src/app.py", PriorityHigh},
		{"lib/util.js", PriorityHigh},
		{"components/Button.tsx", PriorityHigh},
		{"handler.java", PriorityHigh},
		{"lib.rs", PriorityHigh},
		{"main.c", PriorityHigh},
		{"app.cpp", PriorityHigh},
		{"utils.swift", PriorityHigh},
		{"Main.kt", PriorityHigh},
		{"app.rb", PriorityHigh},
		{"server.ex", PriorityHigh},
		{"Program.cs", PriorityHigh},

		// Medium priority - config/scripts
		{"config.yaml", PriorityMedium},
		{"settings.yml", PriorityMedium},
		{"package.json", PriorityMedium},
		{"config.toml", PriorityMedium},
		{"pom.xml", PriorityMedium},
		{"build.sh", PriorityMedium},
		{"deploy.bash", PriorityMedium},
		{"Makefile", PriorityMedium},
		{"Taskfile.yml", PriorityMedium},
		{"Dockerfile", PriorityMedium},
		{"Dockerfile.prod", PriorityMedium},
		{"CMakeLists.txt", PriorityMedium},

		// Low priority - tests/docs
		{"main_test.go", PriorityLow},
		{"app.spec.ts", PriorityLow},
		{"util.test.js", PriorityLow},
		{"README.md", PriorityLow},
		{"CHANGELOG.txt", PriorityLow},
		{"docs/guide.md", PriorityLow},
		{"test/helper.go", PriorityLow},
		{"tests/unit/app.py", PriorityLow},
		{"vendor/lib/util.go", PriorityLow},
		{"node_modules/pkg/index.js", PriorityLow},
		{"api.rst", PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := DetectPriority(tt.path); got != tt.want {
				t.Errorf("DetectPriority(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCountDiffLines(t *testing.T) {
	tests := []struct {
		name    string
		diff    string
		wantAdd int
		wantDel int
	}{
		{
			name:    "empty diff",
			diff:    "",
			wantAdd: 0,
			wantDel: 0,
		},
		{
			name: "only additions",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,5 @@
 package main
+import "fmt"
+import "os"`,
			wantAdd: 2,
			wantDel: 0,
		},
		{
			name: "only deletions",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,3 @@
 package main
-import "fmt"
-import "os"`,
			wantAdd: 0,
			wantDel: 2,
		},
		{
			name: "mixed changes",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,5 @@
 package main
-import "fmt"
+import "log"
+import "os"
-import "io"`,
			wantAdd: 2,
			wantDel: 2,
		},
		{
			name: "ignore header lines",
			diff: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"`,
			wantAdd: 1,
			wantDel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdd, gotDel := CountDiffLines(tt.diff)
			if gotAdd != tt.wantAdd || gotDel != tt.wantDel {
				t.Errorf("CountDiffLines() = (%d, %d), want (%d, %d)",
					gotAdd, gotDel, tt.wantAdd, tt.wantDel)
			}
		})
	}
}

func TestFilterAndTruncateDiffs(t *testing.T) {
	t.Run("filter lock files when >= 5 files", func(t *testing.T) {
		files := []FileDiff{
			{Path: "main.go", Diff: "+line1"},
			{Path: "util.go", Diff: "+line2"},
			{Path: "config.yaml", Diff: "+line3"},
			{Path: "README.md", Diff: "+line4"},
			{Path: "package-lock.json", Diff: "+line5\n+line6\n+line7"},
		}

		result := FilterAndTruncateDiffs(files, 1000)

		// Should have 4 files (lock file removed)
		if len(result) != 4 {
			t.Errorf("expected 4 files, got %d", len(result))
		}

		// Verify lock file is removed
		for _, f := range result {
			if f.Path == "package-lock.json" {
				t.Error("lock file should be filtered out")
			}
		}
	})

	t.Run("keep lock files when < 5 files", func(t *testing.T) {
		files := []FileDiff{
			{Path: "main.go", Diff: "+line1"},
			{Path: "go.sum", Diff: "+line2"},
		}

		result := FilterAndTruncateDiffs(files, 1000)

		if len(result) != 2 {
			t.Errorf("expected 2 files, got %d", len(result))
		}
	})

	t.Run("truncate by priority", func(t *testing.T) {
		// Create files that exceed maxLines
		files := []FileDiff{
			{Path: "main.go", Diff: "+a\n+b\n+c"},            // High: 3 lines
			{Path: "config.yaml", Diff: "+d\n+e"},            // Medium: 2 lines
			{Path: "README.md", Diff: "+f\n+g\n+h\n+i\n+j"}, // Low: 5 lines
		}

		// Only allow 5 lines total - should keep high and medium, truncate low
		result := FilterAndTruncateDiffs(files, 5)

		// Find each file and check
		var mainFile, configFile, readmeFile *FileDiff
		for i := range result {
			switch result[i].Path {
			case "main.go":
				mainFile = &result[i]
			case "config.yaml":
				configFile = &result[i]
			case "README.md":
				readmeFile = &result[i]
			}
		}

		if mainFile == nil || mainFile.Truncated {
			t.Error("main.go should not be truncated")
		}
		if configFile == nil || configFile.Truncated {
			t.Error("config.yaml should not be truncated")
		}
		if readmeFile == nil || !readmeFile.Truncated {
			t.Error("README.md should be truncated")
		}
	})

	t.Run("preserve line counts for truncated files", func(t *testing.T) {
		files := []FileDiff{
			{Path: "main.go", Diff: "+a\n+b\n+c\n-d\n-e"}, // 3 add, 2 del
		}

		result := FilterAndTruncateDiffs(files, 1) // Force truncation

		if result[0].LinesAdd != 3 || result[0].LinesDel != 2 {
			t.Errorf("expected LinesAdd=3, LinesDel=2, got %d, %d",
				result[0].LinesAdd, result[0].LinesDel)
		}
	})

	t.Run("small files preserved first within same priority", func(t *testing.T) {
		files := []FileDiff{
			{Path: "big.go", Diff: "+1\n+2\n+3\n+4\n+5"}, // 5 lines
			{Path: "small.go", Diff: "+a"},               // 1 line
		}

		result := FilterAndTruncateDiffs(files, 3)

		var bigFile, smallFile *FileDiff
		for i := range result {
			switch result[i].Path {
			case "big.go":
				bigFile = &result[i]
			case "small.go":
				smallFile = &result[i]
			}
		}

		if smallFile == nil || smallFile.Truncated {
			t.Error("small.go should not be truncated")
		}
		if bigFile == nil || !bigFile.Truncated {
			t.Error("big.go should be truncated")
		}
	})
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"internal/ui/commit/ai_test.go", true},
		{"src/utils.spec.ts", true},
		{"src/utils.test.js", true},
		{"tests/test_main.py", true},
		{"test/helper.go", true},
		{"internal/ui/commit/ai.go", false},
		{"README.md", false},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsTestFile(tt.path); got != tt.expected {
				t.Errorf("IsTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestIsDocFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"README.md", true},
		{"docs/guide.txt", true},
		{"api.rst", true},
		{"manual.adoc", true},
		{"main.go", false},
		{"config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsDocFile(tt.path); got != tt.expected {
				t.Errorf("IsDocFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"go.mod", true},
		{"go.sum", true},
		{"package.json", true},
		{"config.yaml", true},
		{"settings.yml", true},
		{"config.toml", true},
		{"Makefile", true},
		{"Dockerfile", true},
		{".goreleaser.yaml", true},
		{"main.go", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsConfigFile(tt.path); got != tt.expected {
				t.Errorf("IsConfigFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestDetectCoreFiles(t *testing.T) {
	t.Run("prefers code files", func(t *testing.T) {
		files := []FileDiff{
			{Path: "internal/ui/ai.go", LinesAdd: 100, LinesDel: 50},
			{Path: "internal/ui/ai_test.go", LinesAdd: 50, LinesDel: 20},
			{Path: "README.md", LinesAdd: 10, LinesDel: 5},
			{Path: "internal/llm/client.go", LinesAdd: 80, LinesDel: 30},
			{Path: "go.mod", LinesAdd: 5, LinesDel: 2},
		}

		cores := DetectCoreFiles(files, 2)

		if !cores[0] {
			t.Error("Expected index 0 (ai.go) to be core")
		}
		if !cores[3] {
			t.Error("Expected index 3 (client.go) to be core")
		}
		if cores[1] {
			t.Error("Expected index 1 (ai_test.go) NOT to be core")
		}
		if len(cores) != 2 {
			t.Errorf("Expected 2 core files, got %d", len(cores))
		}
	})

	t.Run("falls back to config and doc when no code", func(t *testing.T) {
		files := []FileDiff{
			{Path: "README.md", LinesAdd: 200, LinesDel: 100},  // doc, 300 lines
			{Path: "go.mod", LinesAdd: 50, LinesDel: 20},       // config, 70 lines
			{Path: "config.yaml", LinesAdd: 30, LinesDel: 10},  // config, 40 lines
			{Path: ".gitignore", LinesAdd: 5, LinesDel: 0},     // other, 5 lines
		}

		cores := DetectCoreFiles(files, 2)

		// Should select config files first (higher priority than doc)
		if !cores[1] {
			t.Error("Expected index 1 (go.mod) to be core")
		}
		if !cores[2] {
			t.Error("Expected index 2 (config.yaml) to be core")
		}
		if cores[0] {
			t.Error("Expected index 0 (README.md) NOT to be core (doc has lower priority)")
		}
		if len(cores) != 2 {
			t.Errorf("Expected 2 core files, got %d", len(cores))
		}
	})

	t.Run("code with fewer lines beats doc with more lines", func(t *testing.T) {
		files := []FileDiff{
			{Path: "README.md", LinesAdd: 500, LinesDel: 200},  // doc, 700 lines
			{Path: "main.go", LinesAdd: 10, LinesDel: 5},       // code, 15 lines
		}

		cores := DetectCoreFiles(files, 1)

		if !cores[1] {
			t.Error("Expected index 1 (main.go) to be core - code priority beats doc")
		}
		if cores[0] {
			t.Error("Expected index 0 (README.md) NOT to be core")
		}
	})
}

func TestSortFilesForCommit(t *testing.T) {
	files := []FileDiff{
		{Path: "README.md", LinesAdd: 10, LinesDel: 5},
		{Path: "internal/ui/ai.go", LinesAdd: 100, LinesDel: 50},
		{Path: ".gitignore", LinesAdd: 1, LinesDel: 0},
		{Path: "internal/ui/ai_test.go", LinesAdd: 50, LinesDel: 20},
		{Path: "go.mod", LinesAdd: 5, LinesDel: 2},
	}

	coreIndices := map[int]bool{1: true} // ai.go is core

	sorted := SortFilesForCommit(files, coreIndices)

	// Expected order: [CORE]ai.go, ai_test.go, go.mod, README.md, .gitignore
	expectedOrder := []string{
		"internal/ui/ai.go",      // [CORE] code
		"internal/ui/ai_test.go", // test
		"go.mod",                 // config
		"README.md",              // doc
		".gitignore",             // other
	}

	for i, expected := range expectedOrder {
		if sorted[i].Path != expected {
			t.Errorf("Position %d: got %s, want %s", i, sorted[i].Path, expected)
		}
	}
}
