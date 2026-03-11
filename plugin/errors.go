package plugin

import "errors"

var (
	// ErrPromptRequired is returned when no prompt is provided
	ErrPromptRequired = errors.New("prompt is required: set PLUGIN_PROMPT or PLUGIN_PROMPT_FILE")

	// ErrGeminiCLINotFound is returned when gemini CLI is not installed
	ErrGeminiCLINotFound = errors.New("gemini CLI not found: ensure gemini is installed and in PATH")

	// ErrCLIExecution is returned when gemini CLI execution fails
	ErrCLIExecution = errors.New("gemini CLI execution failed")

	// ErrOutputParsing is returned when CLI output cannot be parsed
	ErrOutputParsing = errors.New("failed to parse gemini CLI output")

	// ErrTimeout is returned when CLI execution times out
	ErrTimeout = errors.New("gemini CLI execution timed out")

	// ErrFileNotFound is returned when a specified file does not exist
	ErrFileNotFound = errors.New("specified file not found")

	// ErrFileRead is returned when a file cannot be read
	ErrFileRead = errors.New("failed to read file")
)
