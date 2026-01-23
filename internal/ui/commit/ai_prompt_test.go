package commit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/llm"
)

// TestPromptPhase1 tests the file analysis prompt (Phase 1) stability on ONE file.
// Run with: go test -v -run TestPromptPhase1 -count=1 ./internal/ui/commit/
func TestPromptPhase1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping prompt test in short mode")
	}

	rounds := getTestRounds()
	repoPath := getTestRepo()

	t.Logf("Testing Phase 1 (file analysis) with %d rounds", rounds)
	t.Logf("Repository: %s", repoPath)

	// Get staged files from the repository
	files, err := getStagedFiles(repoPath)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("No staged files in repository")
	}

	t.Logf("Found %d staged files", len(files))

	// Create a mock AI model to reuse the prompt building logic
	client := llm.NewClient()

	// Detect core files
	coreIndices := git.DetectCoreFiles(files, 3)
	sortedFiles := git.SortFilesForCommit(files, coreIndices)

	// Rebuild coreIndices for sorted order
	sortedCoreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && coreIndices[origIdx] {
				sortedCoreIndices[i] = true
				break
			}
		}
	}

	// Test the first CORE file (or first file if no CORE)
	testIdx := 0
	for i := range sortedFiles {
		if sortedCoreIndices[i] {
			testIdx = i
			break
		}
	}

	testFile := sortedFiles[testIdx]
	if testFile.Truncated {
		t.Skip("Test file was truncated, skipping")
	}

	t.Logf("Testing file: %s (CORE: %v)", testFile.Path, sortedCoreIndices[testIdx])

	// Build file list context
	fileListContext := buildFileListContext(sortedFiles, sortedCoreIndices)

	// Build the prompt exactly as ai.go does
	marker := ""
	if sortedCoreIndices[testIdx] {
		marker = " [CORE]"
	}
	prompt := fmt.Sprintf("%s\nNow analyze%s: %s\n%s\n\nDescribe what changed in this file, considering its role in the overall commit.",
		fileListContext, marker, testFile.Path, testFile.Diff)

	systemPrompt := consts.LLMDefaultFilePrompt

	// Run multiple rounds
	results := make([]string, rounds)
	var shortResults, goodResults, qualityIssues int

	for i := 0; i < rounds; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := client.Generate(ctx, client.GetModel(), prompt, llm.GenerateOptions{System: systemPrompt})
		cancel()

		if err != nil {
			t.Errorf("Round %d failed: %v", i+1, err)
			continue
		}

		results[i] = result
		wordCount := len(strings.Fields(result))

		// Analyze length
		isShort := wordCount <= 5
		if isShort {
			shortResults++
		} else {
			goodResults++
		}

		// Analyze quality issues
		issues := analyzePhase1Quality(result, testFile.Diff)
		if len(issues) > 0 {
			qualityIssues++
		}

		// Output full result for review
		status := "OK"
		if isShort {
			status = "SHORT"
		}
		if len(issues) > 0 {
			status += " QUALITY:" + strings.Join(issues, ",")
		}
		t.Logf("Round %d [%s] (%d words):\n  %s", i+1, status, wordCount, strings.ReplaceAll(result, "\n", "\n  "))
	}

	t.Logf("\n=== Phase 1 Summary ===")
	t.Logf("Total rounds: %d", rounds)
	t.Logf("Good length (>5 words): %d (%.1f%%)", goodResults, float64(goodResults)*100/float64(rounds))
	t.Logf("Short results (<=5 words): %d (%.1f%%)", shortResults, float64(shortResults)*100/float64(rounds))
	t.Logf("Quality issues: %d (%.1f%%)", qualityIssues, float64(qualityIssues)*100/float64(rounds))

	// Fail if too many short results
	if shortResults > rounds/4 {
		t.Errorf("Too many short results: %d/%d (>25%%)", shortResults, rounds)
	}
}

// TestPromptPhase1AllFiles tests Phase 1 on ALL staged files.
// Run with: go test -v -run TestPromptPhase1AllFiles -count=1 ./internal/ui/commit/
// Set TEST_ROUNDS_PER_FILE to control rounds per file (default 3).
func TestPromptPhase1AllFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping prompt test in short mode")
	}

	roundsPerFile := 3
	if r := os.Getenv("TEST_ROUNDS_PER_FILE"); r != "" {
		fmt.Sscanf(r, "%d", &roundsPerFile)
	}
	repoPath := getTestRepo()

	t.Logf("Testing Phase 1 (file analysis) on ALL files, %d rounds per file", roundsPerFile)
	t.Logf("Repository: %s", repoPath)

	// Get staged files from the repository
	files, err := getStagedFiles(repoPath)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("No staged files in repository")
	}

	t.Logf("Found %d staged files", len(files))

	client := llm.NewClient()

	// Detect core files
	coreIndices := git.DetectCoreFiles(files, 3)
	sortedFiles := git.SortFilesForCommit(files, coreIndices)

	// Rebuild coreIndices for sorted order
	sortedCoreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && coreIndices[origIdx] {
				sortedCoreIndices[i] = true
				break
			}
		}
	}

	// Build file list context (shared for all files)
	fileListContext := buildFileListContext(sortedFiles, sortedCoreIndices)
	systemPrompt := consts.LLMDefaultFilePrompt

	// Stats
	type fileStats struct {
		path          string
		isCore        bool
		truncated     bool
		shortCount    int
		qualityIssues int
		totalRounds   int
		samples       []string
	}
	allStats := make([]fileStats, len(sortedFiles))

	// Test each file
	for fileIdx, testFile := range sortedFiles {
		stats := fileStats{
			path:      testFile.Path,
			isCore:    sortedCoreIndices[fileIdx],
			truncated: testFile.Truncated,
		}

		if testFile.Truncated {
			t.Logf("\n[%d/%d] %s (TRUNCATED - skipped)", fileIdx+1, len(sortedFiles), testFile.Path)
			allStats[fileIdx] = stats
			continue
		}

		marker := ""
		if sortedCoreIndices[fileIdx] {
			marker = " [CORE]"
		}
		prompt := fmt.Sprintf("%s\nNow analyze%s: %s\n%s\n\nDescribe what changed in this file, considering its role in the overall commit.",
			fileListContext, marker, testFile.Path, testFile.Diff)

		t.Logf("\n[%d/%d] %s%s", fileIdx+1, len(sortedFiles), testFile.Path, marker)

		for round := 0; round < roundsPerFile; round++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			result, err := client.Generate(ctx, client.GetModel(), prompt, llm.GenerateOptions{System: systemPrompt})
			cancel()

			if err != nil {
				t.Errorf("  Round %d failed: %v", round+1, err)
				continue
			}

			stats.totalRounds++
			wordCount := len(strings.Fields(result))

			// Analyze length
			isShort := wordCount <= 5
			if isShort {
				stats.shortCount++
			}

			// Analyze quality issues
			issues := analyzePhase1Quality(result, testFile.Diff)
			if len(issues) > 0 {
				stats.qualityIssues++
			}

			// Save sample
			stats.samples = append(stats.samples, result)

			// Output result
			status := "OK"
			if isShort {
				status = "SHORT"
			}
			if len(issues) > 0 {
				status += " QUALITY:" + strings.Join(issues, ",")
			}
			t.Logf("  Round %d [%s] (%d words): %s", round+1, status, wordCount, truncate(result, 80))
		}

		allStats[fileIdx] = stats
	}

	// Print summary
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("=== Phase 1 ALL FILES Summary ===")
	t.Log(strings.Repeat("=", 60))

	var totalFiles, testedFiles, filesWithShort, filesWithQuality int
	var totalRounds, totalShort, totalQuality int

	for _, stats := range allStats {
		totalFiles++
		if stats.truncated {
			continue
		}
		testedFiles++
		totalRounds += stats.totalRounds
		totalShort += stats.shortCount
		totalQuality += stats.qualityIssues

		if stats.shortCount > 0 {
			filesWithShort++
		}
		if stats.qualityIssues > 0 {
			filesWithQuality++
		}
	}

	t.Logf("Total files: %d", totalFiles)
	t.Logf("Tested files: %d (truncated: %d)", testedFiles, totalFiles-testedFiles)
	t.Logf("Total rounds: %d", totalRounds)
	t.Logf("Short results: %d (%.1f%%)", totalShort, float64(totalShort)*100/float64(totalRounds))
	t.Logf("Quality issues: %d (%.1f%%)", totalQuality, float64(totalQuality)*100/float64(totalRounds))
	t.Logf("Files with any short: %d (%.1f%%)", filesWithShort, float64(filesWithShort)*100/float64(testedFiles))
	t.Logf("Files with any quality issue: %d (%.1f%%)", filesWithQuality, float64(filesWithQuality)*100/float64(testedFiles))

	// Show problematic files
	t.Log("\n--- Problematic Files ---")
	hasProblems := false
	for _, stats := range allStats {
		if stats.truncated {
			continue
		}
		if stats.shortCount > 0 || stats.qualityIssues > 0 {
			hasProblems = true
			coreMarker := ""
			if stats.isCore {
				coreMarker = " [CORE]"
			}
			t.Logf("  %s%s: short=%d/%d, quality=%d/%d",
				stats.path, coreMarker,
				stats.shortCount, stats.totalRounds,
				stats.qualityIssues, stats.totalRounds)
			// Show samples
			for i, sample := range stats.samples {
				t.Logf("    Sample %d: %s", i+1, truncate(sample, 70))
			}
		}
	}
	if !hasProblems {
		t.Log("  None! All files passed.")
	}

	// Fail conditions
	shortRate := float64(totalShort) / float64(totalRounds)
	qualityRate := float64(totalQuality) / float64(totalRounds)

	if shortRate > 0.25 {
		t.Errorf("Too many short results: %.1f%% (>25%%)", shortRate*100)
	}
	if qualityRate > 0.25 {
		t.Errorf("Too many quality issues: %.1f%% (>25%%)", qualityRate*100)
	}
}

// TestPromptPhase2 tests the commit message generation prompt (Phase 2) stability.
// Run with: go test -v -run TestPromptPhase2 -count=1 ./internal/ui/commit/
func TestPromptPhase2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping prompt test in short mode")
	}

	rounds := getTestRounds()
	repoPath := getTestRepo()

	t.Logf("Testing Phase 2 (commit message) with %d rounds", rounds)
	t.Logf("Repository: %s", repoPath)

	// Get staged files
	files, err := getStagedFiles(repoPath)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("No staged files in repository")
	}

	t.Logf("Found %d staged files", len(files))

	client := llm.NewClient()

	// First, run Phase 1 to get real summaries
	coreIndices := git.DetectCoreFiles(files, 3)
	sortedFiles := git.SortFilesForCommit(files, coreIndices)

	// Rebuild coreIndices for sorted order
	sortedCoreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && coreIndices[origIdx] {
				sortedCoreIndices[i] = true
				break
			}
		}
	}

	// Get summaries from Phase 1
	summaries := make([]string, len(sortedFiles))
	fileListContext := buildFileListContext(sortedFiles, sortedCoreIndices)

	t.Log("Running Phase 1 to generate summaries...")
	for i, f := range sortedFiles {
		if f.Truncated {
			summaries[i] = fmt.Sprintf("(file changes: +%d/-%d lines, content omitted)", f.LinesAdd, f.LinesDel)
			continue
		}

		marker := ""
		if sortedCoreIndices[i] {
			marker = " [CORE]"
		}
		prompt := fmt.Sprintf("%s\nNow analyze%s: %s\n%s\n\nDescribe what changed in this file, considering its role in the overall commit.",
			fileListContext, marker, f.Path, f.Diff)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := client.Generate(ctx, client.GetModel(), prompt, llm.GenerateOptions{System: consts.LLMDefaultFilePrompt})
		cancel()

		if err != nil {
			t.Logf("Phase 1 failed for %s: %v", f.Path, err)
			summaries[i] = "Analysis failed"
			continue
		}
		summaries[i] = strings.TrimSpace(result)
		t.Logf("  %s: %s", f.Path, truncate(summaries[i], 60))
	}

	// Build Phase 2 prompt exactly as ai.go does
	phase2Prompt := buildCommitPrompt(sortedFiles, summaries, sortedCoreIndices, client.GetLang())
	systemPrompt := getCommitSystemPrompt(client.GetLang())

	t.Log("\n=== Starting Phase 2 Tests ===")

	// Run multiple rounds
	results := make([]string, rounds)
	var goodFormat, badFormat, hasHallucination int

	for i := 0; i < rounds; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := client.Generate(ctx, client.GetModel(), phase2Prompt, llm.GenerateOptions{System: systemPrompt})
		cancel()

		if err != nil {
			t.Errorf("Round %d failed: %v", i+1, err)
			continue
		}

		results[i] = result

		// Analyze result quality
		issues := analyzeCommitMessage(result, sortedFiles)

		if len(issues) == 0 {
			goodFormat++
			t.Logf("Round %d: GOOD\n%s", i+1, indentLines(result, "  "))
		} else {
			badFormat++
			if containsHallucination(issues) {
				hasHallucination++
			}
			t.Logf("Round %d: ISSUES: %v\n%s", i+1, issues, indentLines(result, "  "))
		}
	}

	t.Logf("\n=== Phase 2 Summary ===")
	t.Logf("Total rounds: %d", rounds)
	t.Logf("Good format: %d (%.1f%%)", goodFormat, float64(goodFormat)*100/float64(rounds))
	t.Logf("Bad format: %d (%.1f%%)", badFormat, float64(badFormat)*100/float64(rounds))
	t.Logf("Hallucinations: %d (%.1f%%)", hasHallucination, float64(hasHallucination)*100/float64(rounds))

	// Fail if too many bad results
	if badFormat > rounds/4 {
		t.Errorf("Too many bad format results: %d/%d (>25%%)", badFormat, rounds)
	}
	if hasHallucination > rounds/10 {
		t.Errorf("Too many hallucinations: %d/%d (>10%%)", hasHallucination, rounds)
	}
}

// TestPromptPhase2Only tests Phase 2 with pre-defined summaries (faster).
// Run with: go test -v -run TestPromptPhase2Only -count=1 ./internal/ui/commit/
func TestPromptPhase2Only(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping prompt test in short mode")
	}

	rounds := getTestRounds()
	repoPath := getTestRepo()

	t.Logf("Testing Phase 2 ONLY (with simulated summaries) with %d rounds", rounds)
	t.Logf("Repository: %s", repoPath)

	// Get staged files
	files, err := getStagedFiles(repoPath)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("No staged files in repository")
	}

	t.Logf("Found %d staged files", len(files))

	client := llm.NewClient()

	// Process files
	coreIndices := git.DetectCoreFiles(files, 3)
	sortedFiles := git.SortFilesForCommit(files, coreIndices)

	// Rebuild coreIndices for sorted order
	sortedCoreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && coreIndices[origIdx] {
				sortedCoreIndices[i] = true
				break
			}
		}
	}

	// Generate simple summaries based on diff analysis (simulating Phase 1)
	summaries := make([]string, len(sortedFiles))
	for i, f := range sortedFiles {
		summaries[i] = generateSimpleSummary(f)
		t.Logf("  %s: %s", f.Path, summaries[i])
	}

	// Build Phase 2 prompt
	phase2Prompt := buildCommitPrompt(sortedFiles, summaries, sortedCoreIndices, client.GetLang())
	systemPrompt := getCommitSystemPrompt(client.GetLang())

	t.Log("\n=== Starting Phase 2 Tests ===")

	// Run multiple rounds
	var goodFormat, badFormat, hasHallucination int

	for i := 0; i < rounds; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := client.Generate(ctx, client.GetModel(), phase2Prompt, llm.GenerateOptions{System: systemPrompt})
		cancel()

		if err != nil {
			t.Errorf("Round %d failed: %v", i+1, err)
			continue
		}

		issues := analyzeCommitMessage(result, sortedFiles)

		if len(issues) == 0 {
			goodFormat++
			t.Logf("Round %d: GOOD\n%s", i+1, indentLines(result, "  "))
		} else {
			badFormat++
			if containsHallucination(issues) {
				hasHallucination++
			}
			t.Logf("Round %d: ISSUES: %v\n%s", i+1, issues, indentLines(result, "  "))
		}
	}

	t.Logf("\n=== Phase 2 Summary ===")
	t.Logf("Total rounds: %d", rounds)
	t.Logf("Good format: %d (%.1f%%)", goodFormat, float64(goodFormat)*100/float64(rounds))
	t.Logf("Bad format: %d (%.1f%%)", badFormat, float64(badFormat)*100/float64(rounds))
	t.Logf("Hallucinations: %d (%.1f%%)", hasHallucination, float64(hasHallucination)*100/float64(rounds))
}

// Helper functions

func getTestRounds() int {
	if r := os.Getenv("TEST_ROUNDS"); r != "" {
		var rounds int
		fmt.Sscanf(r, "%d", &rounds)
		if rounds > 0 {
			return rounds
		}
	}
	return 20 // default
}

func getTestRepo() string {
	if r := os.Getenv("TEST_REPO"); r != "" {
		return r
	}
	// Default to current directory
	cwd, _ := os.Getwd()
	// Go up to project root
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "."
}

func getStagedFiles(repoPath string) ([]git.FileDiff, error) {
	// Get diff from the specified repo
	cmd := exec.Command("git", "diff", "--cached", "-U5")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	diff := string(output)
	if diff == "" {
		return nil, nil
	}

	files := git.SplitDiffByFile(diff)
	files = git.FilterAndTruncateDiffs(files, 500)
	return files, nil
}

func buildFileListContext(files []git.FileDiff, coreIndices map[int]bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("This commit changes %d files:\n", len(files)))

	for i, f := range files {
		marker := "      "
		if coreIndices[i] {
			marker = "[CORE]"
		}
		sb.WriteString(fmt.Sprintf("%s %s (+%d/-%d)\n", marker, f.Path, f.LinesAdd, f.LinesDel))
	}

	return sb.String()
}

func buildCommitPrompt(files []git.FileDiff, summaries []string, coreIndices map[int]bool, lang string) string {
	var sb strings.Builder

	// Few-shot examples (same as ai.go)
	switch lang {
	case consts.LLMLangZH:
		sb.WriteString(`示例 1 (单一目的, 1 文件):
输入:
- config.go: 修复超时变量名拼写错误

输出:
fix(config): 修复超时变量名拼写错误

- 修复超时变量名拼写错误

示例 2 (单一目的, 多文件 - body 仍然只需 1 行):
输入:
- cmd/root.go: 更新 import 路径
- internal/service/service.go: 更新 import 路径
- pkg/crypto/processor.go: 更新 import 路径
- go.mod: 更新模块路径

输出:
chore(module): 迁移模块路径

- 迁移模块路径以适配新仓库地址

示例 3 (多个目的 - 每个目的 1 行):
输入:
- auth.go: 添加 JWT 验证
- config.go: 添加认证配置
- logger.go: 修复日志格式问题

输出:
feat(auth): 添加 JWT 认证并修复日志

- 添加 JWT 认证功能
- 修复日志格式问题

输入:
`)
	case consts.LLMLangBilingual:
		sb.WriteString(`Example 1 (single purpose, 1 file):
Input:
- config.go: Fix typo in timeout variable name

Output:
fix(config): correct timeout variable typo (修复超时变量名拼写错误)

- 修复超时变量名拼写错误

Example 2 (single purpose, many files - still 1 line body):
Input:
- cmd/root.go: Update import path
- internal/service/service.go: Update import path
- pkg/crypto/processor.go: Update import path
- go.mod: Update module path

Output:
chore(module): migrate module path (迁移模块路径)

- 迁移模块路径以适配新仓库地址

Example 3 (multiple purposes - 1 line per purpose):
Input:
- auth.go: Add JWT validation
- config.go: Add auth config
- logger.go: Fix log format issue

Output:
feat(auth): add JWT auth and fix logging (添加 JWT 认证并修复日志)

- 添加 JWT 认证功能
- 修复日志格式问题

Input:
`)
	default:
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
	}

	// Add summaries with [CORE] markers
	for i, f := range files {
		summary := summaries[i]
		if summary != "" {
			marker := "      "
			if coreIndices[i] {
				marker = "[CORE]"
			}
			sb.WriteString(fmt.Sprintf("%s %s: %s\n", marker, f.Path, strings.TrimSpace(summary)))
		}
	}

	sb.WriteString("\nOutput (MUST include body with \"- \" lines):")

	return sb.String()
}

func getCommitSystemPrompt(lang string) string {
	switch lang {
	case consts.LLMLangZH:
		return consts.LLMCommitPromptZH
	case consts.LLMLangBilingual:
		return consts.LLMCommitPromptBilingual
	default:
		return consts.LLMCommitPromptEN
	}
}

func generateSimpleSummary(f git.FileDiff) string {
	// Analyze the diff to generate a simple summary
	if f.Truncated {
		return fmt.Sprintf("(file changes: +%d/-%d lines, content omitted)", f.LinesAdd, f.LinesDel)
	}

	// Detect common patterns
	diff := f.Diff
	lowerDiff := strings.ToLower(diff)

	// Check for import path changes
	if strings.Contains(diff, "import") && (strings.Contains(diff, "+") || strings.Contains(diff, "-")) {
		if strings.Contains(lowerDiff, "gitlab") || strings.Contains(lowerDiff, "github") ||
			strings.Contains(lowerDiff, "git.example.com") {
			return "Update import path"
		}
	}

	// Check for module path changes
	if strings.Contains(f.Path, "go.mod") {
		return "Update module path"
	}

	// Check for version changes
	if strings.Contains(f.Path, "go.sum") {
		return "Update dependency checksums"
	}

	// Check for documentation
	ext := strings.ToLower(filepath.Ext(f.Path))
	if ext == ".md" || ext == ".txt" {
		return "Update documentation"
	}

	// Check for config files
	if strings.HasSuffix(f.Path, ".json") || strings.HasSuffix(f.Path, ".yaml") ||
		strings.HasSuffix(f.Path, ".yml") || strings.HasSuffix(f.Path, ".toml") {
		return "Update configuration"
	}

	// Default based on change type
	if f.LinesAdd > f.LinesDel*2 {
		return "Add new functionality"
	} else if f.LinesDel > f.LinesAdd*2 {
		return "Remove unused code"
	}
	return "Update implementation"
}

func analyzeCommitMessage(msg string, files []git.FileDiff) []string {
	var issues []string

	lines := strings.Split(msg, "\n")
	if len(lines) == 0 {
		issues = append(issues, "empty message")
		return issues
	}

	// Check header format
	header := lines[0]
	if !strings.Contains(header, "(") || !strings.Contains(header, "):") {
		issues = append(issues, "invalid header format")
	}

	// Check for body
	hasBody := false
	for _, line := range lines[1:] {
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			hasBody = true
			break
		}
	}
	if !hasBody {
		issues = append(issues, "missing body")
	}

	// Check for hallucination (mentions things not in the input)
	lowerMsg := strings.ToLower(msg)
	hallucinations := []string{
		"animal", "feeding", "schedule", "pet", "zoo",
		"timeout", "retry", "cache", "redis", // unless actually present
		"authentication", "jwt", "oauth", // unless actually present
	}

	// Build a set of words from file paths for context
	fileContext := make(map[string]bool)
	for _, f := range files {
		words := strings.FieldsFunc(f.Path, func(r rune) bool {
			return r == '/' || r == '_' || r == '-' || r == '.'
		})
		for _, w := range words {
			fileContext[strings.ToLower(w)] = true
		}
	}

	for _, h := range hallucinations {
		if strings.Contains(lowerMsg, h) && !fileContext[h] {
			issues = append(issues, fmt.Sprintf("possible hallucination: '%s'", h))
		}
	}

	return issues
}

func containsHallucination(issues []string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, "hallucination") {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// analyzePhase1Quality checks for quality issues in Phase 1 output.
func analyzePhase1Quality(result, diff string) []string {
	var issues []string
	lowerResult := strings.ToLower(result)
	lowerDiff := strings.ToLower(diff)

	// Check for migration direction accuracy
	// If diff shows migration FROM gitlab TO new-domain, result should not say "use GitLab" or "new GitLab"
	if strings.Contains(lowerDiff, "-gitlab") && strings.Contains(lowerDiff, "+git.example.com") {
		if strings.Contains(lowerResult, "use gitlab") || strings.Contains(lowerResult, "to gitlab") ||
			strings.Contains(lowerResult, "new gitlab") {
			issues = append(issues, "wrong-direction")
		}
	}

	// Check for forbidden words (should not appear in output)
	forbiddenWords := []string{
		"readability", "maintainability", "modularity",
		"network", "performance", "connectivity",
	}
	for _, word := range forbiddenWords {
		if strings.Contains(lowerResult, word) && !strings.Contains(lowerDiff, word) {
			issues = append(issues, "forbidden:"+word)
		}
	}

	// Check for vague/speculative phrases
	vaguePhrases := []string{
		"better organization",
		"enhance user experience",
		"easier updates",
		"policy changes",
		"internal policy",
	}
	for _, phrase := range vaguePhrases {
		if strings.Contains(lowerResult, phrase) {
			issues = append(issues, "vague")
			break
		}
	}

	return issues
}

func indentLines(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

// TestFullPipeline runs complete Phase1 + Phase2 pipeline for each round.
// This simulates real usage where each commit generation is independent.
// Run with: TEST_REPO=/path/to/repo TEST_ROUNDS=20 go test -v -run TestFullPipeline -count=1 ./internal/ui/commit/
func TestFullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping prompt test in short mode")
	}

	rounds := getTestRounds()
	repoPath := getTestRepo()

	t.Logf("=== Full Pipeline Test (Phase1 + Phase2) ===")
	t.Logf("Repository: %s", repoPath)
	t.Logf("Rounds: %d", rounds)

	// Get staged files
	files, err := getStagedFiles(repoPath)
	if err != nil {
		t.Fatalf("Failed to get staged files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("No staged files in repository")
	}

	t.Logf("Staged files: %d", len(files))

	client := llm.NewClient()

	// Detect core files and sort
	coreIndices := git.DetectCoreFiles(files, 3)
	sortedFiles := git.SortFilesForCommit(files, coreIndices)

	// Rebuild coreIndices for sorted order
	sortedCoreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && coreIndices[origIdx] {
				sortedCoreIndices[i] = true
				break
			}
		}
	}

	// Build file list context (shared)
	fileListContext := buildFileListContext(sortedFiles, sortedCoreIndices)

	// Results storage
	type roundResult struct {
		round           int
		phase1Summaries []string
		phase1Issues    []string // quality issues in phase 1
		commitMessage   string
		phase2Issues    []string // issues in commit message
		headerLine      string
		bodyLines       []string
	}

	results := make([]roundResult, rounds)

	for round := 0; round < rounds; round++ {
		t.Logf("\n--- Round %d/%d ---", round+1, rounds)
		result := roundResult{round: round + 1}

		// Phase 1: Generate summaries for all files
		summaries := make([]string, len(sortedFiles))
		for i, f := range sortedFiles {
			if f.Truncated {
				summaries[i] = fmt.Sprintf("(file changes: +%d/-%d lines, content omitted)", f.LinesAdd, f.LinesDel)
				continue
			}

			marker := ""
			if sortedCoreIndices[i] {
				marker = " [CORE]"
			}
			prompt := fmt.Sprintf("%s\nNow analyze%s: %s\n%s\n\nDescribe what changed in this file, considering its role in the overall commit.",
				fileListContext, marker, f.Path, f.Diff)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			summary, err := client.Generate(ctx, client.GetModel(), prompt, llm.GenerateOptions{System: consts.LLMDefaultFilePrompt})
			cancel()

			if err != nil {
				summaries[i] = "Analysis failed"
				result.phase1Issues = append(result.phase1Issues, fmt.Sprintf("%s: API error", f.Path))
				continue
			}

			summary = strings.TrimSpace(summary)
			summaries[i] = summary

			// Check Phase 1 quality
			wordCount := len(strings.Fields(summary))
			if wordCount <= 5 {
				result.phase1Issues = append(result.phase1Issues, fmt.Sprintf("%s: too short (%d words)", f.Path, wordCount))
			}

			issues := analyzePhase1Quality(summary, f.Diff)
			for _, issue := range issues {
				result.phase1Issues = append(result.phase1Issues, fmt.Sprintf("%s: %s", f.Path, issue))
			}
		}
		result.phase1Summaries = summaries

		// Phase 2: Generate commit message
		phase2Prompt := buildCommitPrompt(sortedFiles, summaries, sortedCoreIndices, client.GetLang())
		systemPrompt := getCommitSystemPrompt(client.GetLang())

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		commitMsg, err := client.Generate(ctx, client.GetModel(), phase2Prompt, llm.GenerateOptions{System: systemPrompt})
		cancel()

		if err != nil {
			result.phase2Issues = append(result.phase2Issues, "API error")
			results[round] = result
			continue
		}

		result.commitMessage = strings.TrimSpace(commitMsg)

		// Parse and analyze commit message
		lines := strings.Split(result.commitMessage, "\n")
		if len(lines) > 0 {
			result.headerLine = lines[0]
		}
		for _, line := range lines[1:] {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				result.bodyLines = append(result.bodyLines, trimmed)
			}
		}

		// Check Phase 2 quality
		result.phase2Issues = analyzeCommitMessageQuality(result.commitMessage, sortedFiles)

		results[round] = result

		// Print round summary
		p1Status := "OK"
		if len(result.phase1Issues) > 0 {
			p1Status = fmt.Sprintf("ISSUES(%d)", len(result.phase1Issues))
		}
		p2Status := "OK"
		if len(result.phase2Issues) > 0 {
			p2Status = fmt.Sprintf("ISSUES(%d)", len(result.phase2Issues))
		}
		t.Logf("  Phase1: %s, Phase2: %s", p1Status, p2Status)
		t.Logf("  Commit: %s", truncate(result.headerLine, 70))
	}

	// Print detailed summary
	t.Log("\n" + strings.Repeat("=", 70))
	t.Log("=== FULL PIPELINE TEST SUMMARY ===")
	t.Log(strings.Repeat("=", 70))

	var p1OKCount, p2OKCount, fullOKCount int
	var allP1Issues, allP2Issues []string
	headerCounts := make(map[string]int)

	for _, r := range results {
		p1OK := len(r.phase1Issues) == 0
		p2OK := len(r.phase2Issues) == 0

		if p1OK {
			p1OKCount++
		} else {
			allP1Issues = append(allP1Issues, r.phase1Issues...)
		}

		if p2OK {
			p2OKCount++
		} else {
			allP2Issues = append(allP2Issues, r.phase2Issues...)
		}

		if p1OK && p2OK {
			fullOKCount++
		}

		// Track header variations
		headerCounts[r.headerLine]++
	}

	t.Logf("\nOverall Statistics:")
	t.Logf("  Total rounds: %d", rounds)
	t.Logf("  Phase 1 OK: %d (%.1f%%)", p1OKCount, float64(p1OKCount)*100/float64(rounds))
	t.Logf("  Phase 2 OK: %d (%.1f%%)", p2OKCount, float64(p2OKCount)*100/float64(rounds))
	t.Logf("  Full pipeline OK: %d (%.1f%%)", fullOKCount, float64(fullOKCount)*100/float64(rounds))

	// Header consistency
	t.Logf("\nCommit Header Variations (%d unique):", len(headerCounts))
	type headerStat struct {
		header string
		count  int
	}
	var headerStats []headerStat
	for h, c := range headerCounts {
		headerStats = append(headerStats, headerStat{h, c})
	}
	sort.Slice(headerStats, func(i, j int) bool {
		return headerStats[i].count > headerStats[j].count
	})
	for i, hs := range headerStats {
		if i >= 5 {
			t.Logf("  ... and %d more variations", len(headerStats)-5)
			break
		}
		t.Logf("  [%d] %s", hs.count, truncate(hs.header, 60))
	}

	// Phase 1 issues breakdown
	if len(allP1Issues) > 0 {
		t.Logf("\nPhase 1 Issues (%d total):", len(allP1Issues))
		issueCounts := make(map[string]int)
		for _, issue := range allP1Issues {
			// Extract issue type
			parts := strings.SplitN(issue, ": ", 2)
			if len(parts) == 2 {
				issueCounts[parts[1]]++
			} else {
				issueCounts[issue]++
			}
		}
		for issue, count := range issueCounts {
			t.Logf("  [%d] %s", count, issue)
		}
	}

	// Phase 2 issues breakdown
	if len(allP2Issues) > 0 {
		t.Logf("\nPhase 2 Issues (%d total):", len(allP2Issues))
		issueCounts := make(map[string]int)
		for _, issue := range allP2Issues {
			issueCounts[issue]++
		}
		for issue, count := range issueCounts {
			t.Logf("  [%d] %s", count, issue)
		}
	}

	// Show sample outputs
	t.Log("\n--- Sample Outputs ---")
	shown := 0
	for _, r := range results {
		if shown >= 3 {
			break
		}
		t.Logf("\nRound %d:", r.round)
		t.Logf("  %s", indentLines(r.commitMessage, "  "))
		shown++
	}

	// Fail conditions
	p1Rate := float64(p1OKCount) / float64(rounds)
	p2Rate := float64(p2OKCount) / float64(rounds)

	if p1Rate < 0.75 {
		t.Errorf("Phase 1 success rate too low: %.1f%% (<75%%)", p1Rate*100)
	}
	if p2Rate < 0.75 {
		t.Errorf("Phase 2 success rate too low: %.1f%% (<75%%)", p2Rate*100)
	}
}

// analyzeCommitMessageQuality checks commit message for various quality issues.
func analyzeCommitMessageQuality(msg string, files []git.FileDiff) []string {
	var issues []string

	lines := strings.Split(msg, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		issues = append(issues, "empty message")
		return issues
	}

	header := lines[0]

	// Check header format: type(scope): subject
	if !strings.Contains(header, "(") || !strings.Contains(header, "):") {
		issues = append(issues, "invalid header format (missing type(scope): pattern)")
	}

	// Check header length
	if len(header) > 72 {
		issues = append(issues, fmt.Sprintf("header too long (%d > 72 chars)", len(header)))
	}

	// Check for valid type
	validTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore", "perf", "build"}
	hasValidType := false
	lowerHeader := strings.ToLower(header)
	for _, t := range validTypes {
		if strings.HasPrefix(lowerHeader, t+"(") {
			hasValidType = true
			break
		}
	}
	if !hasValidType {
		issues = append(issues, "invalid commit type")
	}

	// Check for body (must have at least one "- " line)
	hasBody := false
	bodyLineCount := 0
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			hasBody = true
			bodyLineCount++
		}
	}
	if !hasBody {
		issues = append(issues, "missing body (no lines starting with '- ')")
	}

	// Check for excessive body lines (should summarize, not enumerate)
	if bodyLineCount > 5 {
		issues = append(issues, fmt.Sprintf("too many body lines (%d > 5, should summarize)", bodyLineCount))
	}

	// Check for file enumeration in body (bad practice)
	fileEnumCount := 0
	for _, line := range lines[1:] {
		for _, f := range files {
			if strings.Contains(line, filepath.Base(f.Path)) {
				fileEnumCount++
				break
			}
		}
	}
	if fileEnumCount > 3 && len(files) > 5 {
		issues = append(issues, fmt.Sprintf("body enumerates files (%d file names found, should summarize)", fileEnumCount))
	}

	// Check for hallucination
	lowerMsg := strings.ToLower(msg)
	hallucinations := []string{
		"animal", "feeding", "schedule", "pet", "zoo",
		"authentication", "jwt", "oauth", "redis", "cache",
		"database", "migration", "user", "login",
	}

	// Build context from files
	fileContext := strings.ToLower(strings.Join(func() []string {
		var paths []string
		for _, f := range files {
			paths = append(paths, f.Path)
			paths = append(paths, f.Diff)
		}
		return paths
	}(), " "))

	for _, h := range hallucinations {
		if strings.Contains(lowerMsg, h) && !strings.Contains(fileContext, h) {
			issues = append(issues, fmt.Sprintf("possible hallucination: '%s'", h))
		}
	}

	return issues
}
