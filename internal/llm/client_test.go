package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mritd/gitflow-toolkit/v3/config"
	"github.com/mritd/gitflow-toolkit/v3/consts"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"host only", "localhost:11434", "https://localhost:11434"},
		{"with http", "http://localhost:11434", "http://localhost:11434"},
		{"with https", "https://ollama.example.com", "https://ollama.example.com"},
		{"with trailing slash", "http://localhost:11434/", "http://localhost:11434"},
		{"ip address", "192.168.1.100:11434", "https://192.168.1.100:11434"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHost(tt.input)
			if got != tt.want {
				t.Errorf("normalizeHost(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		wantProvider Provider
		wantPath     string
	}{
		{"groq", "https://api.groq.com", ProviderGroq, consts.LLMPathGroq},
		{"openai", "https://api.openai.com", ProviderOpenAI, consts.LLMPathOpenAI},
		{"deepseek", "https://api.deepseek.com", ProviderOpenAI, consts.LLMPathOpenAI},
		{"mistral", "https://api.mistral.ai", ProviderOpenAI, consts.LLMPathOpenAI},
		{"openrouter", "https://openrouter.ai", ProviderOpenRouter, consts.LLMPathOpenRouter},
		{"unknown", "https://custom-llm.example.com", ProviderOpenAI, consts.LLMPathOpenAI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, path := detectProvider(tt.host)
			if provider != tt.wantProvider {
				t.Errorf("detectProvider(%q) provider = %q, want %q", tt.host, provider, tt.wantProvider)
			}
			if path != tt.wantPath {
				t.Errorf("detectProvider(%q) path = %q, want %q", tt.host, path, tt.wantPath)
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	// NOTE: These tests may be affected by ~/.gitconfig settings.
	// Skip if any gitconfig LLM settings are set.
	if config.GetString(config.GitConfigLLMAPIKey, "") != "" ||
		config.GetString(config.GitConfigLLMModel, "") != "" ||
		config.GetString(config.GitConfigLLMAPIHost, "") != "" {
		t.Skip("Skipping: gitconfig has LLM settings")
	}

	t.Run("defaults to Ollama when no API key", func(t *testing.T) {
		c := NewClient()
		if c.provider != ProviderOllama {
			t.Errorf("provider = %q, want %q", c.provider, ProviderOllama)
		}
		if c.host != consts.LLMHostOllama {
			t.Errorf("host = %q, want %q", c.host, consts.LLMHostOllama)
		}
		if c.model != consts.LLMModelOllama {
			t.Errorf("model = %q, want %q", c.model, consts.LLMModelOllama)
		}
		if c.timeout != consts.LLMDefaultRequestTimeout {
			t.Errorf("timeout = %v, want %v", c.timeout, consts.LLMDefaultRequestTimeout)
		}
	})
}

func TestGenerate_Ollama(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/api/generate" {
				t.Errorf("path = %s, want /api/generate", r.URL.Path)
			}

			var req ollamaRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}
			if req.Model != "test-model" {
				t.Errorf("model = %s, want test-model", req.Model)
			}

			resp := ollamaResponse{Response: "  generated text  "}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		c := &Client{
			provider: ProviderOllama,
			host:     server.URL,
			apiPath:  consts.LLMPathOllama,
			timeout:  10 * time.Second,
			retries:  0,
		}

		result, err := c.Generate(context.Background(), "test-model", "test prompt")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if result != "generated text" {
			t.Errorf("Generate() = %q, want %q", result, "generated text")
		}
	})

	t.Run("retry on failure", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp := ollamaResponse{Response: "success"}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		c := &Client{
			provider: ProviderOllama,
			host:     server.URL,
			apiPath:  consts.LLMPathOllama,
			timeout:  10 * time.Second,
			retries:  2,
		}

		result, err := c.Generate(context.Background(), "test-model", "test prompt")
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if result != "success" {
			t.Errorf("Generate() = %q, want %q", result, "success")
		}
		if attempts != 3 {
			t.Errorf("attempts = %d, want 3", attempts)
		}
	})
}

func TestGenerate_OpenAI(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/openai/v1/chat/completions" {
				t.Errorf("path = %s, want /openai/v1/chat/completions", r.URL.Path)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Authorization = %s, want Bearer test-key", auth)
			}

			var req openAIRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}
			if req.Model != "test-model" {
				t.Errorf("model = %s, want test-model", req.Model)
			}
			if len(req.Messages) != 2 {
				t.Errorf("messages count = %d, want 2", len(req.Messages))
			}

			resp := openAIResponse{
				Choices: []openAIChoice{{
					Message: openAIMessage{Role: "assistant", Content: "  generated text  "},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		c := &Client{
			provider: ProviderGroq,
			host:     server.URL,
			apiPath:  consts.LLMPathGroq,
			apiKey:   "test-key",
			timeout:  10 * time.Second,
			retries:  0,
		}

		opt := GenerateOptions{System: "system prompt"}
		result, err := c.Generate(context.Background(), "test-model", "test prompt", opt)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if result != "generated text" {
			t.Errorf("Generate() = %q, want %q", result, "generated text")
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIResponse{
				Error: &openAIError{Message: "rate limit exceeded"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		c := &Client{
			provider: ProviderGroq,
			host:     server.URL,
			apiPath:  consts.LLMPathGroq,
			apiKey:   "test-key",
			timeout:  10 * time.Second,
			retries:  0,
		}

		_, err := c.Generate(context.Background(), "test-model", "test prompt")
		if err == nil {
			t.Fatal("Generate() expected error, got nil")
		}
	})
}

func TestGetDiffContext(t *testing.T) {
	// NOTE: This test may be affected by ~/.gitconfig settings.
	if ctx := config.GetString(config.GitConfigLLMDiffContext, ""); ctx != "" {
		t.Skip("Skipping: gitconfig has llm-diff-context set")
	}

	t.Run("default", func(t *testing.T) {
		if got := GetDiffContext(); got != consts.LLMDefaultDiffContext {
			t.Errorf("GetDiffContext() = %d, want %d", got, consts.LLMDefaultDiffContext)
		}
	})
}
