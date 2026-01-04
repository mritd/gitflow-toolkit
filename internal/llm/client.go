// Package llm provides a unified HTTP client for LLM APIs.
// Supports Ollama (local) and OpenAI-compatible APIs (OpenRouter, Groq, OpenAI).
//
// Configuration priority: gitconfig > environment variable > default value
//
// Example gitconfig (~/.gitconfig):
//
//	[gitflow]
//	    llm-api-key = sk-or-v1-xxxxx
//	    llm-model = mistralai/devstral-2512:free
//	    llm-temperature = 0.3
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mritd/gitflow-toolkit/v3/config"
	"github.com/mritd/gitflow-toolkit/v3/consts"
)

// Provider represents the LLM provider type.
type Provider string

const (
	ProviderOllama     Provider = "ollama"
	ProviderGroq       Provider = "groq"
	ProviderOpenRouter Provider = "openrouter"
	ProviderOpenAI     Provider = "openai"
)

// Client is an LLM API client supporting multiple providers.
type Client struct {
	provider           Provider
	host               string
	apiKey             string
	timeout            time.Duration
	retries            int
	lang               string
	model              string
	temperature        float64
	filePrompt         string
	commitPromptEN     string
	commitPromptZH     string
	commitPromptBiling string
}

// GenerateOptions configures a generation request.
type GenerateOptions struct {
	System      string
	Temperature float64
}

// NewClient creates a new LLM client from gitconfig.
//
// Provider selection:
//   - If API key is set, uses OpenAI-compatible API (OpenRouter by default)
//   - Otherwise, uses local Ollama
func NewClient() *Client {
	// Get API key from gitconfig
	apiKey := config.GetString(config.GitConfigLLMAPIKey, "")

	// Determine provider and defaults based on API key presence
	provider := ProviderOllama
	defaultHost := consts.LLMDefaultOllamaHost
	defaultModel := consts.LLMDefaultOllamaModel

	if apiKey != "" {
		provider = ProviderOpenRouter
		defaultHost = consts.LLMDefaultOpenRouterHost
		defaultModel = consts.LLMDefaultOpenRouterModel
	}

	// Get host
	host := config.GetString(config.GitConfigLLMHost, "")
	if host == "" {
		host = defaultHost
	} else {
		host = normalizeHost(host, defaultHost)
	}

	// Detect provider from host
	if apiKey != "" {
		if strings.Contains(host, "groq.com") {
			provider = ProviderGroq
		} else if strings.Contains(host, "openai.com") {
			provider = ProviderOpenAI
		}
	}

	// Get model
	model := config.GetString(config.GitConfigLLMModel, defaultModel)

	// Get other settings
	timeout := config.GetInt(config.GitConfigLLMTimeout, consts.LLMDefaultTimeout)
	retries := config.GetInt(config.GitConfigLLMRetries, consts.LLMDefaultRetries)
	temperature := config.GetFloat(config.GitConfigLLMTemperature, consts.LLMDefaultTemperature)

	// Get language (validate value)
	lang := config.GetString(config.GitConfigLLMLang, consts.LLMDefaultLang)
	switch lang {
	case consts.LLMLangEN, consts.LLMLangZH, consts.LLMLangBilingual:
		// valid
	default:
		lang = consts.LLMDefaultLang
	}

	// Get custom prompts
	filePrompt := config.GetString(config.GitConfigLLMFilePrompt, "")
	commitPromptEN := config.GetString(config.GitConfigLLMCommitPromptEN, "")
	commitPromptZH := config.GetString(config.GitConfigLLMCommitPromptZH, "")
	commitPromptBiling := config.GetString(config.GitConfigLLMCommitPromptBiling, "")

	return &Client{
		provider:           provider,
		host:               host,
		apiKey:             apiKey,
		timeout:            time.Duration(timeout) * time.Second,
		retries:            retries,
		lang:               lang,
		model:              model,
		temperature:        temperature,
		filePrompt:         filePrompt,
		commitPromptEN:     commitPromptEN,
		commitPromptZH:     commitPromptZH,
		commitPromptBiling: commitPromptBiling,
	}
}

// normalizeHost ensures the host has a proper scheme.
func normalizeHost(host, defaultHost string) string {
	if host == "" {
		return defaultHost
	}
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimSuffix(host, "/")
	}
	return "https://" + strings.TrimSuffix(host, "/")
}

// Generate calls the LLM API to generate text.
// Returns the generated text or error after retries exhausted.
func (c *Client) Generate(ctx context.Context, model, prompt string, opts ...GenerateOptions) (string, error) {
	var opt GenerateOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Use client's default temperature if not specified
	if opt.Temperature == 0 {
		opt.Temperature = c.temperature
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		var result string
		var err error

		if c.provider == ProviderOllama {
			result, err = c.doGenerateOllama(ctx, model, prompt, opt)
		} else {
			result, err = c.doGenerateOpenAI(ctx, model, prompt, opt)
		}

		if err == nil {
			return result, nil
		}
		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", c.retries+1, lastErr)
}

// Ollama API types
type ollamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	System  string         `json:"system,omitempty"`
	Stream  bool           `json:"stream"`
	Options *ollamaOptions `json:"options,omitempty"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (c *Client) doGenerateOllama(ctx context.Context, model, prompt string, opt GenerateOptions) (string, error) {
	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		System: opt.System,
		Stream: false,
	}

	if opt.Temperature > 0 {
		reqBody.Options = &ollamaOptions{Temperature: opt.Temperature}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return strings.TrimSpace(result.Response), nil
}

// OpenAI-compatible API types (works with OpenRouter, Groq, OpenAI)
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIError struct {
	Message string `json:"message"`
}

func (c *Client) doGenerateOpenAI(ctx context.Context, model, prompt string, opt GenerateOptions) (string, error) {
	messages := make([]openAIMessage, 0, 2)

	if opt.System != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: opt.System})
	}
	messages = append(messages, openAIMessage{Role: "user", Content: prompt})

	reqBody := openAIRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 1024,
	}

	if opt.Temperature > 0 {
		reqBody.Temperature = opt.Temperature
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build endpoint URL based on provider
	// Groq uses /openai/v1/..., OpenRouter/OpenAI use /api/v1/...
	var endpoint string
	switch c.provider {
	case ProviderGroq:
		endpoint = c.host + "/openai/v1/chat/completions"
	default:
		endpoint = c.host + "/api/v1/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// GetModel returns the configured model name.
func (c *Client) GetModel() string {
	return c.model
}

// GetLang returns the configured language.
func (c *Client) GetLang() string {
	return c.lang
}

// GetFilePrompt returns the custom file analysis prompt, or empty string for default.
func (c *Client) GetFilePrompt() string {
	return c.filePrompt
}

// GetCommitPrompt returns the custom commit generation prompt for the specified language,
// or empty string to use the default prompt.
func (c *Client) GetCommitPrompt(lang string) string {
	switch lang {
	case consts.LLMLangZH:
		return c.commitPromptZH
	case consts.LLMLangBilingual:
		return c.commitPromptBiling
	default:
		return c.commitPromptEN
	}
}

// GetContext returns the configured diff context lines.
func GetContext() int {
	return config.GetInt(config.GitConfigLLMContext, consts.LLMDefaultContext)
}

// GetConcurrency returns the configured concurrency limit for parallel file analysis.
func GetConcurrency() int {
	return config.GetInt(config.GitConfigLLMConcurrency, consts.LLMDefaultConcurrency)
}
