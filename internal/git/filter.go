package git

import "path/filepath"

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
