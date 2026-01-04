package config

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GitConfigSection name for gitflow-toolkit settings.
const GitConfigSection = "gitflow"

// GitConfig keys.
const (
	GitConfigLLMAPIKey                = "llm-api-key"
	GitConfigLLMAPIHost               = "llm-api-host"
	GitConfigLLMAPIPath               = "llm-api-path"
	GitConfigLLMModel                 = "llm-model"
	GitConfigLLMTemperature           = "llm-temperature"
	GitConfigLLMRequestTimeout        = "llm-request-timeout"
	GitConfigLLMMaxRetries            = "llm-max-retries"
	GitConfigLLMOutputLang            = "llm-output-lang"
	GitConfigLLMDiffContext           = "llm-diff-context"
	GitConfigLLMMaxConcurrency        = "llm-max-concurrency"
	GitConfigLLMFileAnalysisPrompt    = "llm-file-analysis-prompt"
	GitConfigLLMCommitPromptEN        = "llm-commit-prompt-en"
	GitConfigLLMCommitPromptZH        = "llm-commit-prompt-zh"
	GitConfigLLMCommitPromptBilingual = "llm-commit-prompt-bilingual"
	GitConfigLuckyCommitPrefix        = "lucky-commit-prefix"
	GitConfigSSHStrictHostKey         = "ssh-strict-host-key"
	GitConfigBranchAutoDetect         = "branch-auto-detect"
)

// gitConfig runs git config --get and returns the value.
func gitConfig(key string) string {
	fullKey := GitConfigSection + "." + key
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("git.exe", "config", "--get", fullKey)
	} else {
		cmd = exec.Command("git", "config", "--get", fullKey)
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetString returns a config value from gitconfig, or default if not set.
func GetString(gitKey, defaultVal string) string {
	if val := gitConfig(gitKey); val != "" {
		return val
	}
	return defaultVal
}

// GetInt returns an int config value from gitconfig, or default if not set.
func GetInt(gitKey string, defaultVal int) int {
	if val := gitConfig(gitKey); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			return v
		}
	}
	return defaultVal
}

// GetFloat returns a float64 config value from gitconfig, or default if not set.
func GetFloat(gitKey string, defaultVal float64) float64 {
	if val := gitConfig(gitKey); val != "" {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			return v
		}
	}
	return defaultVal
}

// GetBool returns a bool config value from gitconfig, or default if not set.
func GetBool(gitKey string, defaultVal bool) bool {
	if val := strings.ToLower(gitConfig(gitKey)); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

// GetDuration returns a time.Duration config value from gitconfig, or default if not set.
// Supports Go duration format (e.g., "30s", "2m", "1h30m").
func GetDuration(gitKey string, defaultVal time.Duration) time.Duration {
	if val := gitConfig(gitKey); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
