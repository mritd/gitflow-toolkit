package git

import (
	"reflect"
	"testing"
)

func TestSplitDiffByFile(t *testing.T) {
	tests := []struct {
		name string
		diff string
		want []FileDiff
	}{
		{
			name: "empty diff",
			diff: "",
			want: nil,
		},
		{
			name: "single file",
			diff: `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"`,
			want: []FileDiff{
				{
					Path: "main.go",
					Diff: `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"`,
				},
			},
		},
		{
			name: "multiple files",
			diff: `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
+line1
diff --git a/pkg/file2.go b/pkg/file2.go
--- a/pkg/file2.go
+++ b/pkg/file2.go
+line2`,
			want: []FileDiff{
				{
					Path: "file1.go",
					Diff: `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
+line1`,
				},
				{
					Path: "pkg/file2.go",
					Diff: `diff --git a/pkg/file2.go b/pkg/file2.go
--- a/pkg/file2.go
+++ b/pkg/file2.go
+line2`,
				},
			},
		},
		{
			name: "nested path",
			diff: `diff --git a/internal/config/config.go b/internal/config/config.go
+changes`,
			want: []FileDiff{
				{
					Path: "internal/config/config.go",
					Diff: `diff --git a/internal/config/config.go b/internal/config/config.go
+changes`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitDiffByFile(tt.diff)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitDiffByFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "simple file",
			line: "diff --git a/main.go b/main.go",
			want: "main.go",
		},
		{
			name: "nested path",
			line: "diff --git a/internal/config/config.go b/internal/config/config.go",
			want: "internal/config/config.go",
		},
		{
			name: "file with spaces",
			line: "diff --git a/path with spaces/file.go b/path with spaces/file.go",
			want: "path with spaces/file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFilePath(tt.line)
			if got != tt.want {
				t.Errorf("extractFilePath(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}
