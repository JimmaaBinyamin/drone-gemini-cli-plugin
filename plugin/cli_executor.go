package plugin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CLIExecutor executes gemini CLI commands
type CLIExecutor struct {
	config *Config
}

// ExecutionResult holds the result of a CLI execution
type ExecutionResult struct {
	RawOutput string
	Response  *CLIResponse
	ExitCode  int
}

// NewCLIExecutor creates a new CLI executor
func NewCLIExecutor(config *Config) *CLIExecutor {
	return &CLIExecutor{config: config}
}

// CheckGeminiCLI verifies that gemini CLI is installed
func (e *CLIExecutor) CheckGeminiCLI() error {
	cmd := exec.Command("gemini", "--version")
	if err := cmd.Run(); err != nil {
		return ErrGeminiCLINotFound
	}
	return nil
}

// Execute runs the gemini CLI with the configured options
func (e *CLIExecutor) Execute(prompt string, stdinInput string) (*ExecutionResult, error) {
	// Build command arguments
	args := e.buildArgs(prompt)

	if e.config.Debug {
		fmt.Printf("[DEBUG] Executing: gemini %s\n", strings.Join(args, " "))
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.Timeout)*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, "gemini", args...)
	cmd.Dir = e.config.Target

	// Set up input/output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Handle stdin input
	if stdinInput != "" {
		cmd.Stdin = strings.NewReader(stdinInput)
	}

	// Set environment with authentication
	cmd.Env = e.buildEnv()

	if e.config.Debug {
		fmt.Println("[DEBUG] Environment variables set for authentication")
	}

	// Execute
	err := cmd.Run()

	result := &ExecutionResult{
		RawOutput: stdout.String(),
		ExitCode:  0,
	}

	// Handle errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}

		// If we got some output, try to parse it (might contain error details)
		if stdout.Len() > 0 {
			if e.config.Debug {
				fmt.Printf("[DEBUG] CLI stderr: %s\n", stderr.String())
			}
		} else {
			return nil, fmt.Errorf("%w: %v (stderr: %s)", ErrCLIExecution, err, stderr.String())
		}
	}

	// Parse output based on format
	parser := NewOutputParser(e.config.Debug)

	switch e.config.OutputFormat {
	case "json":
		response, parseErr := parser.ParseJSON(result.RawOutput)
		if parseErr != nil {
			return result, parseErr
		}
		result.Response = response

	case "stream-json":
		_, response, parseErr := parser.ParseStreamJSON(result.RawOutput)
		if parseErr != nil {
			return result, parseErr
		}
		result.Response = response

	default: // text
		result.Response = parser.ParseText(result.RawOutput)
	}

	// Check for API errors in response
	if result.Response != nil && result.Response.Error != nil {
		return result, fmt.Errorf("%w: %s - %s",
			ErrCLIExecution,
			result.Response.Error.Type,
			result.Response.Error.Message)
	}

	return result, nil
}

// buildEnv constructs environment variables including authentication
func (e *CLIExecutor) buildEnv() []string {
	// Start with current environment
	env := os.Environ()

	authMode := e.config.DetectAuthMode()

	switch authMode {
	case AuthModeGeminiAPIKey:
		// Gemini API Key (Google AI Studio)
		// Use GEMINI_API_KEY for headless mode
		env = append(env, fmt.Sprintf("GEMINI_API_KEY=%s", e.config.APIKey))
		if e.config.Debug {
			fmt.Println("[DEBUG] Using Gemini API Key authentication (Google AI Studio)")
		}

	case AuthModeVertexAI:
		// Vertex AI authentication
		// IMPORTANT: Must set GOOGLE_GENAI_USE_VERTEXAI=true for Vertex AI mode
		env = append(env, "GOOGLE_GENAI_USE_VERTEXAI=true")

		// Set project and location
		if e.config.GCPProject != "" {
			env = append(env, fmt.Sprintf("GOOGLE_CLOUD_PROJECT=%s", e.config.GCPProject))
		}
		if e.config.GCPLocation != "" {
			env = append(env, fmt.Sprintf("GOOGLE_CLOUD_LOCATION=%s", e.config.GCPLocation))
		}

		// Use APIKey as GOOGLE_API_KEY for Vertex AI
		if e.config.APIKey != "" {
			env = append(env, fmt.Sprintf("GOOGLE_API_KEY=%s", e.config.APIKey))
			if e.config.Debug {
				fmt.Printf("[DEBUG] Using Vertex AI with Google API Key (Project: %s)\n", e.config.GCPProject)
			}
		}

		// Service Account credentials (alternative to API Key)
		if e.config.GCPCredentials != "" {
			credPath := e.config.GCPCredentials
			// Check if it's JSON content rather than a path
			if strings.HasPrefix(strings.TrimSpace(e.config.GCPCredentials), "{") {
				// It's JSON content, write to temp file
				tmpFile, err := os.CreateTemp("", "gcp-credentials-*.json")
				if err == nil {
					tmpFile.WriteString(e.config.GCPCredentials)
					tmpFile.Close()
					credPath = tmpFile.Name()
					if e.config.Debug {
						fmt.Printf("[DEBUG] Wrote credentials to temp file: %s\n", credPath)
					}
				}
			}
			// Make path absolute if needed
			if !filepath.IsAbs(credPath) {
				if absPath, err := filepath.Abs(credPath); err == nil {
					credPath = absPath
				}
			}
			env = append(env, fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", credPath))
			if e.config.Debug {
				fmt.Printf("[DEBUG] Using Vertex AI with Service Account (Project: %s)\n", e.config.GCPProject)
			}
		}

	default:
		if e.config.Debug {
			fmt.Println("[DEBUG] No authentication configured, relying on existing credentials")
		}
	}

	return env
}

// buildArgs constructs command line arguments based on configuration
func (e *CLIExecutor) buildArgs(prompt string) []string {
	var args []string

	// Required: prompt
	args = append(args, "--prompt", prompt)

	// Output format
	if e.config.OutputFormat != "" {
		args = append(args, "--output-format", e.config.OutputFormat)
	}

	// Model selection
	if e.config.Model != "" {
		args = append(args, "--model", e.config.Model)
	}

	// YOLO mode (auto-approve all actions)
	if e.config.Yolo {
		args = append(args, "--yolo")
	}

	// Approval mode
	if e.config.ApprovalMode != "" {
		args = append(args, "--approval-mode", e.config.ApprovalMode)
	}

	// Include directories
	if e.config.IncludeDirs != "" {
		args = append(args, "--include-directories", e.config.IncludeDirs)
	}

	// Debug mode
	if e.config.Debug {
		args = append(args, "--debug")
	}

	return args
}
