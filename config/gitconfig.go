package config

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GitConfigSection name for gitflow-toolkit settings.
const GitConfigSection = "gitflow"

// GitConfig keys.
const (
	GitConfigLLMAPIKey                = "llm-api-key"
	GitConfigLLMHost                  = "llm-host"
	GitConfigLLMModel                 = "llm-model"
	GitConfigLLMTemperature           = "llm-temperature"
	GitConfigLLMTimeout               = "llm-timeout"
	GitConfigLLMRetries               = "llm-retries"
	GitConfigLLMLang                  = "llm-lang"
	GitConfigLLMContext               = "llm-context"
	GitConfigLLMConcurrency           = "llm-concurrency"
	GitConfigLLMFilePrompt            = "llm-file-prompt"
	GitConfigLLMCommitPromptEN        = "llm-commit-prompt-en"
	GitConfigLLMCommitPromptZH        = "llm-commit-prompt-zh"
	GitConfigLLMCommitPromptBilingual = "llm-commit-prompt-bilingual"
	GitConfigLuckyCommit              = "lucky-commit"
	GitConfigSSHStrictHost            = "ssh-strict-host-key"
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
