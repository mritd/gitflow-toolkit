package git

import (
	"os"
	"os/exec"
	"testing"

	"github.com/mritd/gitflow-toolkit/v3/config"
)

func TestValidateLuckyPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		want    string
		wantErr bool
	}{
		{"valid lowercase", "abc123", "abc123", false},
		{"valid uppercase converts", "ABC123", "abc123", false},
		{"valid mixed case", "AbC123", "abc123", false},
		{"empty string", "", "", true},
		{"too long", "12345678901234567", "", true},
		{"exactly 12 chars", "1234567890ab", "1234567890ab", false},
		{"invalid chars", "xyz123", "", true},
		{"invalid with space", "abc 123", "", true},
		{"valid all zeros", "0000000000000000", "0000000000000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateLuckyPrefix(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLuckyPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateLuckyPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckLuckyCommit(t *testing.T) {
	// Save original PATH
	origPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", origPath) }()

	// Test with empty PATH (lucky_commit not found)
	_ = os.Setenv("PATH", "")
	err := CheckLuckyCommit()
	if err == nil {
		t.Error("CheckLuckyCommit() should return error when not in PATH")
	}

	// Test with lucky_commit in PATH (if available)
	_ = os.Setenv("PATH", origPath)
	if _, lookErr := exec.LookPath("lucky_commit"); lookErr == nil {
		err = CheckLuckyCommit()
		if err != nil {
			t.Errorf("CheckLuckyCommit() unexpected error: %v", err)
		}
	}
}

func TestGetLuckyPrefix(t *testing.T) {
	// NOTE: GetLuckyPrefix reads from gitconfig, which cannot be easily mocked.
	// Skip if gitconfig has lucky-commit set.
	if prefix := config.GetString(config.GitConfigLuckyCommit, ""); prefix != "" {
		t.Skip("Skipping: gitconfig has lucky-commit set")
	}

	// Test returns empty when gitconfig not set
	prefix := GetLuckyPrefix()
	if prefix != "" {
		t.Errorf("GetLuckyPrefix() = %v, want empty", prefix)
	}
}
