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
