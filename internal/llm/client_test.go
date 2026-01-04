package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/config"
)

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		defaultHost string
		want        string
	}{
		{"empty with default", "", "http://default:1234", "http://default:1234"},
		{"empty no default", "", "", ""},
		{"host only", "localhost:11434", "", "https://localhost:11434"},
		{"with http", "http://localhost:11434", "", "http://localhost:11434"},
		{"with https", "https://ollama.example.com", "", "https://ollama.example.com"},
		{"with trailing slash", "http://localhost:11434/", "", "http://localhost:11434"},
		{"ip address", "192.168.1.100:11434", "", "https://192.168.1.100:11434"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHost(tt.input, tt.defaultHost)
			if got != tt.want {
				t.Errorf("normalizeHost(%q, %q) = %q, want %q", tt.input, tt.defaultHost, got, tt.want)
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	// NOTE: These tests may be affected by ~/.gitconfig settings.
	// Skip if any gitconfig LLM settings are set.
	if config.GetString(config.GitConfigLLMAPIKey, "") != "" ||
		config.GetString(config.GitConfigLLMModel, "") != "" ||
		config.GetString(config.GitConfigLLMHost, "") != "" {
		t.Skip("Skipping: gitconfig has LLM settings")
	}

	t.Run("defaults to Ollama when no API key", func(t *testing.T) {
		c := NewClient()
		if c.provider != ProviderOllama {
			t.Errorf("provider = %q, want %q", c.provider, ProviderOllama)
		}
		if c.host != consts.LLMDefaultOllamaHost {
			t.Errorf("host = %q, want %q", c.host, consts.LLMDefaultOllamaHost)
		}
		if c.model != consts.LLMDefaultOllamaModel {
			t.Errorf("model = %q, want %q", c.model, consts.LLMDefaultOllamaModel)
		}
		if c.timeout != time.Duration(consts.LLMDefaultTimeout)*time.Second {
			t.Errorf("timeout = %v, want %v", c.timeout, time.Duration(consts.LLMDefaultTimeout)*time.Second)
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

func TestGetContext(t *testing.T) {
	// NOTE: This test may be affected by ~/.gitconfig settings.
	if ctx := config.GetString(config.GitConfigLLMContext, ""); ctx != "" {
		t.Skip("Skipping: gitconfig has llm-context set")
	}

	t.Run("default", func(t *testing.T) {
		if got := GetContext(); got != consts.LLMDefaultContext {
			t.Errorf("GetContext() = %d, want %d", got, consts.LLMDefaultContext)
		}
	})
}
