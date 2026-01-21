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
