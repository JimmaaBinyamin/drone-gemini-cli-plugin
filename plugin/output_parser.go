package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CLIResponse represents the JSON output from gemini CLI
type CLIResponse struct {
	Response string    `json:"response"`
	Stats    *CLIStats `json:"stats"`
	Error    *CLIError `json:"error"`
}

// CLIStats holds usage statistics from gemini CLI
type CLIStats struct {
	Models map[string]ModelStats `json:"models"`
	Tools  ToolStats             `json:"tools"`
	Files  FileStats             `json:"files"`
}

// ModelStats holds per-model statistics
type ModelStats struct {
	API    APIStats   `json:"api"`
	Tokens TokenStats `json:"tokens"`
}

// APIStats holds API call statistics
type APIStats struct {
	TotalRequests  int `json:"totalRequests"`
	TotalErrors    int `json:"totalErrors"`
	TotalLatencyMs int `json:"totalLatencyMs"`
}

// TokenStats holds token usage statistics
type TokenStats struct {
	Prompt     int `json:"prompt"`
	Candidates int `json:"candidates"`
	Total      int `json:"total"`
	Cached     int `json:"cached"`
	Thoughts   int `json:"thoughts"`
	Tool       int `json:"tool"`
}

// ToolStats holds tool usage statistics
type ToolStats struct {
	TotalCalls      int                   `json:"totalCalls"`
	TotalSuccess    int                   `json:"totalSuccess"`
	TotalFail       int                   `json:"totalFail"`
	TotalDurationMs int                   `json:"totalDurationMs"`
	TotalDecisions  ToolDecisions         `json:"totalDecisions"`
	ByName          map[string]ToolDetail `json:"byName"`
}

// ToolDecisions holds decision statistics
type ToolDecisions struct {
	Accept     int `json:"accept"`
	Reject     int `json:"reject"`
	Modify     int `json:"modify"`
	AutoAccept int `json:"auto_accept"`
}

// ToolDetail holds per-tool statistics
type ToolDetail struct {
	Count      int           `json:"count"`
	Success    int           `json:"success"`
	Fail       int           `json:"fail"`
	DurationMs int           `json:"durationMs"`
	Decisions  ToolDecisions `json:"decisions"`
}

// FileStats holds file modification statistics
type FileStats struct {
	TotalLinesAdded   int `json:"totalLinesAdded"`
	TotalLinesRemoved int `json:"totalLinesRemoved"`
}

// CLIError represents an error from gemini CLI
type CLIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// StreamEvent represents a single event from stream-json output
type StreamEvent struct {
	Type      string `json:"type"` // init, message, tool_use, tool_result, error, result
	Timestamp string `json:"timestamp"`

	// init event fields
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`

	// message event fields
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	Delta   bool   `json:"delta,omitempty"`

	// tool_use event fields
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolID     string                 `json:"tool_id,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// tool_result event fields
	Status string `json:"status,omitempty"`
	Output string `json:"output,omitempty"`

	// result event fields
	Stats *CLIStats `json:"stats,omitempty"`
}

// OutputParser parses gemini CLI output
type OutputParser struct {
	debug bool
}

// NewOutputParser creates a new output parser
func NewOutputParser(debug bool) *OutputParser {
	return &OutputParser{debug: debug}
}

// ParseJSON parses JSON format output from gemini CLI
func (p *OutputParser) ParseJSON(output string) (*CLIResponse, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("%w: empty output", ErrOutputParsing)
	}

	var response CLIResponse
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		if p.debug {
			fmt.Printf("[DEBUG] Failed to parse JSON: %v\n", err)
			fmt.Printf("[DEBUG] Raw output: %s\n", output)
		}
		return nil, fmt.Errorf("%w: %v", ErrOutputParsing, err)
	}

	return &response, nil
}

// ParseStreamJSON parses stream-json format output (JSONL)
func (p *OutputParser) ParseStreamJSON(output string) ([]StreamEvent, *CLIResponse, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var events []StreamEvent
	var finalResponse *CLIResponse

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event StreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			if p.debug {
				fmt.Printf("[DEBUG] Failed to parse stream event: %v\n", err)
			}
			continue
		}

		events = append(events, event)

		// Extract final result
		if event.Type == "result" {
			finalResponse = &CLIResponse{
				Stats: event.Stats,
			}
		}

		// Extract final message content
		if event.Type == "message" && event.Role == "assistant" && !event.Delta {
			if finalResponse == nil {
				finalResponse = &CLIResponse{}
			}
			finalResponse.Response = event.Content
		}
	}

	return events, finalResponse, nil
}

// ParseText handles text format output (returns as-is)
func (p *OutputParser) ParseText(output string) *CLIResponse {
	return &CLIResponse{
		Response: strings.TrimSpace(output),
	}
}
