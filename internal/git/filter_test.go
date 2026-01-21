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
