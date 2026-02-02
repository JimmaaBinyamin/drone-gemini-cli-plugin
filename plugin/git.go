package plugin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommitInfo holds information about a git commit
type CommitInfo struct {
	SHA       string
	Author    string
	Email     string
	Message   string
	Timestamp string
}

// GitAnalyzer handles git operations for code review
type GitAnalyzer struct {
	repoPath string
	debug    bool
}

// NewGitAnalyzer creates a new git analyzer
func NewGitAnalyzer(repoPath string, debug bool) *GitAnalyzer {
	return &GitAnalyzer{
		repoPath: repoPath,
		debug:    debug,
	}
}

// DetectCommitSHA detects the commit SHA from environment or git
func (g *GitAnalyzer) DetectCommitSHA(configSHA string) string {
	// Priority 1: Explicit configuration
	if configSHA != "" {
		return configSHA
	}

	// Priority 2: Drone CI environment variables
	droneEnvVars := []string{
		"DRONE_COMMIT_SHA",
		"DRONE_COMMIT",
		"CI_COMMIT_SHA",
		"GITHUB_SHA",
		"GITLAB_CI_COMMIT_SHA",
	}

	for _, envVar := range droneEnvVars {
		if sha := os.Getenv(envVar); sha != "" {
			if g.debug {
				fmt.Printf("[DEBUG] Detected commit SHA from %s: %s\n", envVar, sha)
			}
			return sha
		}
	}

	// Priority 3: Get HEAD from git
	sha, err := g.runGitCommand("rev-parse", "HEAD")
	if err == nil && sha != "" {
		if g.debug {
			fmt.Printf("[DEBUG] Detected commit SHA from git HEAD: %s\n", sha)
		}
		return strings.TrimSpace(sha)
	}

	return ""
}

// GetCommitInfo retrieves information about a commit
func (g *GitAnalyzer) GetCommitInfo(sha string) (*CommitInfo, error) {
	if sha == "" {
		sha = "HEAD"
	}

	// Get commit info in a single command
	format := "%H%n%an%n%ae%n%s%n%ci"
	output, err := g.runGitCommand("log", "-1", "--format="+format, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 5 {
		return nil, fmt.Errorf("unexpected git log output")
	}

	return &CommitInfo{
		SHA:       lines[0],
		Author:    lines[1],
		Email:     lines[2],
		Message:   lines[3],
		Timestamp: lines[4],
	}, nil
}

// GetChangedFiles returns list of files changed in a commit
func (g *GitAnalyzer) GetChangedFiles(sha string) ([]string, error) {
	if sha == "" {
		sha = "HEAD"
	}

	// Get list of changed files
	output, err := g.runGitCommand("diff-tree", "--no-commit-id", "--name-only", "-r", sha)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	var result []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			result = append(result, f)
		}
	}

	return result, nil
}

// GetCommitDiff returns the diff of a commit
func (g *GitAnalyzer) GetCommitDiff(sha string) (string, error) {
	if sha == "" {
		sha = "HEAD"
	}

	// Get the diff with some context
	output, err := g.runGitCommand("diff", sha+"^.."+sha, "--unified=3")
	if err != nil {
		// Try without parent (for initial commit)
		output, err = g.runGitCommand("show", sha, "--format=", "--unified=3")
		if err != nil {
			return "", fmt.Errorf("failed to get commit diff: %w", err)
		}
	}

	return output, nil
}

// GetDiffStats returns a summary of changes in a commit
func (g *GitAnalyzer) GetDiffStats(sha string) (string, error) {
	if sha == "" {
		sha = "HEAD"
	}

	output, err := g.runGitCommand("diff", sha+"^.."+sha, "--stat")
	if err != nil {
		output, err = g.runGitCommand("show", sha, "--format=", "--stat")
		if err != nil {
			return "", fmt.Errorf("failed to get diff stats: %w", err)
		}
	}

	return output, nil
}

// IsGitRepository checks if the path is a git repository
func (g *GitAnalyzer) IsGitRepository() bool {
	_, err := g.runGitCommand("rev-parse", "--git-dir")
	return err == nil
}

// BuildGitContext builds a context string with git information
func (g *GitAnalyzer) BuildGitContext(sha string) (string, error) {
	var context strings.Builder

	// Get commit info
	commitInfo, err := g.GetCommitInfo(sha)
	if err != nil {
		return "", err
	}

	context.WriteString("=== Git Commit Information ===\n")
	context.WriteString(fmt.Sprintf("Commit: %s\n", commitInfo.SHA[:12]))
	context.WriteString(fmt.Sprintf("Author: %s <%s>\n", commitInfo.Author, commitInfo.Email))
	context.WriteString(fmt.Sprintf("Date: %s\n", commitInfo.Timestamp))
	context.WriteString(fmt.Sprintf("Message: %s\n", commitInfo.Message))
	context.WriteString("\n")

	// Get changed files
	changedFiles, err := g.GetChangedFiles(sha)
	if err == nil && len(changedFiles) > 0 {
		context.WriteString("=== Changed Files ===\n")
		for _, f := range changedFiles {
			context.WriteString(fmt.Sprintf("- %s\n", f))
		}
		context.WriteString("\n")
	}

	// Get diff stats
	stats, err := g.GetDiffStats(sha)
	if err == nil && stats != "" {
		context.WriteString("=== Change Statistics ===\n")
		context.WriteString(stats)
		context.WriteString("\n")
	}

	// Get actual diff (truncate if too long)
	diff, err := g.GetCommitDiff(sha)
	if err == nil && diff != "" {
		context.WriteString("=== Commit Diff ===\n")
		// Limit diff size to ~50KB to leave room for code context
		if len(diff) > 50000 {
			context.WriteString(diff[:50000])
			context.WriteString("\n... [diff truncated due to size] ...\n")
		} else {
			context.WriteString(diff)
		}
		context.WriteString("\n")
	}

	return context.String(), nil
}

// runGitCommand executes a git command and returns the output
func (g *GitAnalyzer) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), stderr.String())
	}

	return stdout.String(), nil
}
