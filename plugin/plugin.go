package plugin

import (
	"fmt"
	"strings"
)

// Plugin represents the drone-gemini-cli-plugin
type Plugin struct {
	config Config
}

// New creates a new plugin instance
func New(cfg Config) *Plugin {
	return &Plugin{
		config: cfg,
	}
}

// Exec runs the plugin and returns any error encountered
func (p *Plugin) Exec() error {
	// Validate configuration
	if err := p.config.Validate(); err != nil {
		return err
	}

	// Display configuration summary
	p.displayConfig()

	// Create CLI executor
	executor := NewCLIExecutor(&p.config)

	// Check if gemini CLI is available
	if err := executor.CheckGeminiCLI(); err != nil {
		return err
	}

	// Build prompt with optional git context
	prompt := p.config.Prompt
	stdinInput := p.config.StdinInput

	// If git diff mode is enabled, add git context
	if p.config.GitDiff {
		gitContext, err := p.buildGitContext()
		if err != nil {
			fmt.Printf("Warning: failed to build git context: %v\n", err)
		} else {
			// Append git context to stdin input
			if stdinInput != "" {
				stdinInput = stdinInput + "\n\n" + gitContext
			} else {
				stdinInput = gitContext
			}
		}
	}

	// Execute CLI
	fmt.Println("Executing gemini CLI...")
	fmt.Println()

	result, err := executor.Execute(prompt, stdinInput)
	if err != nil {
		return err
	}

	// Display results
	p.displayResult(result)

	return nil
}

// buildGitContext builds git context for the prompt
func (p *Plugin) buildGitContext() (string, error) {
	analyzer := NewGitAnalyzer(p.config.Target, p.config.Debug)

	if !analyzer.IsGitRepository() {
		return "", fmt.Errorf("not a git repository: %s", p.config.Target)
	}

	sha := analyzer.DetectCommitSHA(p.config.GitCommitSHA)
	if sha == "" {
		return "", fmt.Errorf("could not detect commit SHA")
	}

	return analyzer.BuildGitContext(sha)
}

// displayConfig shows the current configuration
func (p *Plugin) displayConfig() {
	fmt.Println()
	fmt.Println("--- Configuration ---")
	fmt.Printf("Target: %s\n", p.config.Target)
	fmt.Printf("Model: %s\n", p.config.Model)
	fmt.Printf("Prompt: %s\n", truncateString(p.config.Prompt, 100))
	fmt.Printf("Output Format: %s\n", p.config.OutputFormat)
	fmt.Printf("Timeout: %ds\n", p.config.Timeout)

	// Display authentication mode
	authMode := p.config.DetectAuthMode()
	switch authMode {
	case AuthModeGeminiAPIKey:
		fmt.Println("Auth: Gemini API Key (Google AI Studio)")
	case AuthModeVertexAI:
		fmt.Printf("Auth: Vertex AI (Project: %s, Location: %s)\n", p.config.GCPProject, p.config.GCPLocation)
	default:
		fmt.Println("Auth: None (using existing credentials)")
	}

	if p.config.Yolo {
		fmt.Println("YOLO Mode: enabled ⚡")
	}

	if p.config.ApprovalMode != "" {
		fmt.Printf("Approval Mode: %s\n", p.config.ApprovalMode)
	}

	if p.config.IncludeDirs != "" {
		fmt.Printf("Include Dirs: %s\n", p.config.IncludeDirs)
	}

	if p.config.GitDiff {
		fmt.Println("Git Diff: enabled")
	}

	if p.config.Debug {
		fmt.Println("Debug: enabled")
	}

	fmt.Println()
}

// displayResult shows the execution result
func (p *Plugin) displayResult(result *ExecutionResult) {
	if result == nil || result.Response == nil {
		fmt.Println("No response received")
		return
	}

	fmt.Println("=== AI Response ===")
	fmt.Println()
	fmt.Println(result.Response.Response)

	// Display statistics
	if result.Response.Stats != nil {
		fmt.Print(FormatStats(result.Response.Stats))
	}

	// Display exit code if non-zero
	if result.ExitCode != 0 {
		fmt.Printf("\n⚠️  Exit Code: %d\n", result.ExitCode)
	}
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, maxLen int) string {
	// Handle multi-line strings
	if idx := strings.Index(s, "\n"); idx != -1 && idx < maxLen {
		return s[:idx] + "..."
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
