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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mritd/gitflow-toolkit/v3/config"
	"github.com/mritd/gitflow-toolkit/v3/consts"
)

// debugLogPath is the path to the LLM debug log file.
var debugLogPath = filepath.Join(os.TempDir(), "gitflow-llm-debug.log")

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
	provider              Provider
	host                  string
	apiPath               string
	apiKey                string
	timeout               time.Duration
	retries               int
	lang                  string
	model                 string
	temperature           float64
	filePrompt            string
	commitPromptEN        string
	commitPromptZH        string
	commitPromptBilingual string
	debug                 bool
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
//
// API path resolution:
//  1. User-defined llm-api-path takes highest priority
//  2. Auto-detect from host for known providers
//  3. Fall back to OpenAI-compatible path (/v1/chat/completions)
func NewClient() *Client {
	// Get API key from gitconfig
	apiKey := config.GetString(config.GitConfigLLMAPIKey, "")

	// Determine provider and defaults based on API key presence
	provider := ProviderOllama
	defaultHost := consts.LLMHostOllama
	defaultModel := consts.LLMModelOllama
	defaultPath := consts.LLMPathOllama

	if apiKey != "" {
		provider = ProviderOpenRouter
		defaultHost = consts.LLMHostOpenRouter
		defaultModel = consts.LLMModelOpenRouter
		defaultPath = consts.LLMPathOpenRouter
	}

	// Get host and normalize
	host := config.GetString(config.GitConfigLLMAPIHost, "")
	if host == "" {
		host = defaultHost
	} else {
		host = normalizeHost(host)
	}

	// Detect provider from host and set appropriate API path
	if apiKey != "" {
		provider, defaultPath = detectProvider(host)
	}

	// Get user-defined API path (highest priority)
	apiPath := config.GetString(config.GitConfigLLMAPIPath, "")
	if apiPath == "" {
		apiPath = defaultPath
	}

	// Get model
	model := config.GetString(config.GitConfigLLMModel, defaultModel)

	// Get other settings
	timeout := config.GetDuration(config.GitConfigLLMRequestTimeout, consts.LLMDefaultRequestTimeout)
	retries := config.GetInt(config.GitConfigLLMMaxRetries, consts.LLMDefaultRetries)
	temperature := config.GetFloat(config.GitConfigLLMTemperature, consts.LLMDefaultTemperature)

	// Get language (validate value)
	lang := config.GetString(config.GitConfigLLMOutputLang, consts.LLMDefaultLang)
	switch lang {
	case consts.LLMLangEN, consts.LLMLangZH, consts.LLMLangBilingual:
		// valid
	default:
		lang = consts.LLMDefaultLang
	}

	// Get custom prompts
	filePrompt := config.GetString(config.GitConfigLLMFileAnalysisPrompt, "")
	commitPromptEN := config.GetString(config.GitConfigLLMCommitPromptEN, "")
	commitPromptZH := config.GetString(config.GitConfigLLMCommitPromptZH, "")
	commitPromptBilingual := config.GetString(config.GitConfigLLMCommitPromptBilingual, "")

	// Get debug flag
	debug := config.GetBool(config.GitConfigLLMAPIDebug, false)

	return &Client{
		provider:              provider,
		host:                  host,
		apiPath:               apiPath,
		apiKey:                apiKey,
		timeout:               timeout,
		retries:               retries,
		lang:                  lang,
		model:                 model,
		temperature:           temperature,
		filePrompt:            filePrompt,
		commitPromptEN:        commitPromptEN,
		commitPromptZH:        commitPromptZH,
		commitPromptBilingual: commitPromptBilingual,
		debug:                 debug,
	}
}

// detectProvider detects the LLM provider from host and returns the provider type and default API path.
// For unknown hosts, returns OpenRouter provider with OpenAI-compatible path.
func detectProvider(host string) (Provider, string) {
	switch {
	case strings.Contains(host, "groq.com"):
		return ProviderGroq, consts.LLMPathGroq
	case strings.Contains(host, "openai.com"):
		return ProviderOpenAI, consts.LLMPathOpenAI
	case strings.Contains(host, "deepseek.com"):
		return ProviderOpenAI, consts.LLMPathOpenAI // DeepSeek uses OpenAI-compatible API
	case strings.Contains(host, "mistral.ai"):
		return ProviderOpenAI, consts.LLMPathOpenAI // Mistral uses OpenAI-compatible API
	case strings.Contains(host, "openrouter.ai"):
		return ProviderOpenRouter, consts.LLMPathOpenRouter
	default:
		// Unknown provider, assume OpenAI-compatible API
		return ProviderOpenAI, consts.LLMPathOpenAI
	}
}

// normalizeHost ensures the host has a proper scheme and no trailing slash.
func normalizeHost(host string) string {
	host = strings.TrimSuffix(host, "/")
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "https://" + host
}

// debugLog writes debug information to the log file if debug mode is enabled.
// Log file location is printed to stderr on first write.
// Uses file locking to prevent interleaved writes from concurrent requests.
func (c *Client) debugLog(format string, args ...interface{}) {
	if !c.debug {
		return
	}

	f, err := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(f, "[%s] %s\n", timestamp, msg)
}

// debugLogRequest logs a complete API request/response pair atomically.
func (c *Client) debugLogRequest(provider, endpoint string, reqBody []byte, respStatus int, respBody []byte, err error) {
	if !c.debug {
		return
	}

	f, err2 := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("=", 80)))
	sb.WriteString(fmt.Sprintf("[%s] %s Request\n", timestamp, provider))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 80)))
	sb.WriteString(fmt.Sprintf("POST %s\n\n", endpoint))

	// Pretty print request JSON
	var prettyReq bytes.Buffer
	if json.Indent(&prettyReq, reqBody, "", "  ") == nil {
		sb.WriteString(fmt.Sprintf("Request:\n%s\n\n", prettyReq.String()))
	} else {
		sb.WriteString(fmt.Sprintf("Request:\n%s\n\n", string(reqBody)))
	}

	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(fmt.Sprintf("Response Status: %d\n", respStatus))
		// Pretty print response JSON
		var prettyResp bytes.Buffer
		if json.Indent(&prettyResp, respBody, "", "  ") == nil {
			sb.WriteString(fmt.Sprintf("Response:\n%s\n", prettyResp.String()))
		} else {
			sb.WriteString(fmt.Sprintf("Response:\n%s\n", string(respBody)))
		}
	}
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("=", 80)))

	_, _ = f.WriteString(sb.String())
}

// GetDebugLogPath returns the path to the debug log file.
func GetDebugLogPath() string {
	return debugLogPath
}

// ClearDebugLog clears the debug log file if debug mode is enabled.
func (c *Client) ClearDebugLog() {
	if !c.debug {
		return
	}
	_ = os.Remove(debugLogPath)
}

// IsDebug returns whether debug mode is enabled.
func (c *Client) IsDebug() bool {
	return c.debug
}

// DebugLog writes a debug message to the log file.
// This is a public method for external callers to log debug info.
func (c *Client) DebugLog(format string, args ...interface{}) {
	c.debugLog(format, args...)
}

// DebugLogSection writes a formatted section to the debug log.
func (c *Client) DebugLogSection(title string, content string) {
	if !c.debug {
		return
	}

	f, err := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("-", 80)))
	sb.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, title))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 80)))
	sb.WriteString(content)
	sb.WriteString("\n")

	_, _ = f.WriteString(sb.String())
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

// Ollama chat API types (preferred for models with chat templates like deepseek-coder-v2)
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

// doGenerateOllama uses the /api/chat endpoint for better compatibility with model templates.
// Models like deepseek-coder-v2 have chat templates that expect User/Assistant format.
func (c *Client) doGenerateOllama(ctx context.Context, model, prompt string, opt GenerateOptions) (string, error) {
	messages := make([]ollamaMessage, 0, 2)

	if opt.System != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: opt.System})
	}
	messages = append(messages, ollamaMessage{Role: "user", Content: prompt})

	reqBody := ollamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	if opt.Temperature > 0 {
		reqBody.Options = &ollamaOptions{Temperature: opt.Temperature}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.host + consts.LLMPathOllama

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.debugLogRequest("Ollama", endpoint, body, 0, nil, err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Log request/response atomically
	c.debugLogRequest("Ollama", endpoint, body, resp.StatusCode, respBody, nil)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return strings.TrimSpace(result.Message.Content), nil
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

	endpoint := c.host + c.apiPath

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.debugLogRequest("OpenAI", endpoint, body, 0, nil, err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Log request/response atomically
	c.debugLogRequest("OpenAI", endpoint, body, resp.StatusCode, respBody, nil)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
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
		return c.commitPromptBilingual
	default:
		return c.commitPromptEN
	}
}

// GetDiffContext returns the configured diff context lines.
func GetDiffContext() int {
	return config.GetInt(config.GitConfigLLMDiffContext, consts.LLMDefaultDiffContext)
}

// GetConcurrency returns the configured concurrency limit for parallel file analysis.
func GetConcurrency() int {
	return config.GetInt(config.GitConfigLLMMaxConcurrency, consts.LLMDefaultConcurrency)
}
