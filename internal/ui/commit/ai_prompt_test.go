package commit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/llm"
)

// testRepo represents a test repository with its diff file
type testRepo struct {
	Name     string
	DiffFile string
}

// TestAICommitGeneration tests the AI commit message generation with multiple repositories.
// Run with: go test -v -run TestAICommitGeneration ./internal/ui/commit/
func TestAICommitGeneration(t *testing.T) {
	// Get testdata directory
	_, filename, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(filename), "testdata")

	// Define test repositories
	repos := []testRepo{
		{Name: "gitflow-toolkit", DiffFile: filepath.Join(testdataDir, "gitflow_toolkit.diff")},
		{Name: "psoss_bak", DiffFile: filepath.Join(testdataDir, "psoss_bak.diff")},
	}

	// Create LLM client
	client := llm.NewClient()

	// Test each repository
	for _, repo := range repos {
		t.Run(repo.Name, func(t *testing.T) {
			testRepoCommitGeneration(t, client, repo)
		})
	}
}

func testRepoCommitGeneration(t *testing.T, client *llm.Client, repo testRepo) {
	// Read diff from file
	diffBytes, err := os.ReadFile(repo.DiffFile)
	if err != nil {
		t.Fatalf("Failed to read diff file %s: %v", repo.DiffFile, err)
	}
	diff := string(diffBytes)
	if diff == "" {
		t.Skip("Empty diff file")
	}

	// Split diff by file and apply filtering (same as production code)
	files := git.SplitDiffByFile(diff)
	if len(files) == 0 {
		t.Skip("No files in diff")
	}

	// Apply lock file filtering and truncation (maxLines=500 as in production)
	files = git.FilterAndTruncateDiffs(files, 500)

	t.Logf("Testing %s with %d files", repo.Name, len(files))

	// Run 20 iterations per repository
	iterations := 20
	results := make([]string, 0, iterations)
	failedCount := 0

	for i := 0; i < iterations; i++ {
		t.Logf("\n=== %s Iteration %d ===", repo.Name, i+1)

		// Phase 1: Analyze each file
		summaries := make([]string, len(files))
		for j, file := range files {
			summary, err := analyzeFile(client, file)
			if err != nil {
				t.Logf("  [%s] Error: %v", file.Path, err)
				continue
			}
			summaries[j] = summary
			t.Logf("  [%s] %s", file.Path, summary)
		}

		// Phase 2: Generate commit message
		commitMsg, err := generateCommitMessage(client, files, summaries)
		if err != nil {
			t.Errorf("Failed to generate commit message: %v", err)
			continue
		}

		t.Logf("\n--- Generated Commit Message ---\n%s\n", commitMsg)
		results = append(results, commitMsg)

		// Validate format
		if !validateCommitMessage(t, commitMsg) {
			failedCount++
		}
	}

	// Summary
	t.Logf("\n=== %s Summary ===", repo.Name)
	t.Logf("Ran %d iterations, %d failed", len(results), failedCount)
	if failedCount > 0 {
		t.Errorf("%s: Failed %d/%d iterations", repo.Name, failedCount, len(results))
	}
}

func analyzeFile(client *llm.Client, file git.FileDiff) (string, error) {
	prompt := fmt.Sprintf(`File: %s
Diff:
%s

Summarize in max 10 words (start with verb):`, file.Path, file.Diff)

	opt := llm.GenerateOptions{
		System: consts.LLMDefaultFilePrompt,
	}
	return client.Generate(context.Background(), client.GetModel(), prompt, opt)
}

func generateCommitMessage(client *llm.Client, files []git.FileDiff, summaries []string) (string, error) {
	var sb strings.Builder

	// Few-shot examples (English)
	sb.WriteString(`Example 1 (single purpose, 1 file):
Input:
- config.go: Fix typo in timeout variable name

Output:
fix(config): correct timeout variable typo

- Correct timeout variable typo

Example 2 (IMPORTANT - many files with SAME action, body is ONE summary line, NOT file list):
Input:
- cmd/root.go: Update import path
- internal/service/service.go: Update import path
- pkg/crypto/processor.go: Update import path
- pkg/utils/helper.go: Update import path
- main.go: Update import path
- go.mod: Update module path
(... 20+ more files with same change ...)

Output:
chore(module): migrate module paths

- Migrate module paths for new repository location

Example 3 (multiple purposes - 1 line per purpose):
Input:
- cache.go: Add Redis cache support
- config.go: Add cache config
- metrics.go: Fix counter reset bug

Output:
feat(cache): add Redis cache and fix metrics

- Add Redis cache support
- Fix counter reset bug

Input:
`)

	for i, summary := range summaries {
		if summary != "" {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", files[i].Path, strings.TrimSpace(summary)))
		}
	}

	sb.WriteString("\nOutput (MUST include body with \"- \" lines):")

	opt := llm.GenerateOptions{
		System: consts.LLMCommitPromptEN,
	}
	return client.Generate(context.Background(), client.GetModel(), sb.String(), opt)
}

func validateCommitMessage(t *testing.T, msg string) bool {
	lines := strings.Split(strings.TrimSpace(msg), "\n")
	valid := true

	// Check header exists
	if len(lines) == 0 {
		t.Error("Validation: Empty commit message")
		return false
	}

	header := lines[0]

	// Check header format: type(scope): subject
	if !strings.Contains(header, "(") || !strings.Contains(header, ")") || !strings.Contains(header, ":") {
		t.Errorf("Validation: Invalid header format: %s", header)
		valid = false
	}

	// Check for body (should have blank line + body lines starting with "- ")
	hasBody := false
	for i, line := range lines {
		if i > 1 && strings.HasPrefix(line, "- ") {
			hasBody = true
			break
		}
	}

	if !hasBody {
		t.Error("Validation: Missing body (no lines starting with '- ')")
		valid = false
	}

	// Check for common issues
	if strings.Contains(msg, "```") {
		t.Error("Validation: Contains markdown code block")
		valid = false
	}
	// Check for literal URLs (AI should describe action, not content)
	if strings.Contains(msg, "://") && (strings.Contains(msg, ".com") || strings.Contains(msg, ".cn")) {
		t.Error("Validation: Contains literal URL (should describe action, not content)")
		valid = false
	}

	// Check for common hallucinations (features from few-shot examples that shouldn't appear)
	hallucinations := []string{"JWT", "jwt", "authentication", "security", "OAuth", "oauth", "log format", "Redis", "cache"}
	msgLower := strings.ToLower(msg)
	for _, h := range hallucinations {
		if strings.Contains(msgLower, strings.ToLower(h)) {
			t.Errorf("Validation: Possible hallucination detected - '%s' copied from examples", h)
			valid = false
			break
		}
	}

	return valid
}
