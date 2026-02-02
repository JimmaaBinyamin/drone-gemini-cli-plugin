package plugin

import (
	"strings"
	"testing"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantResp string
	}{
		{
			name:    "empty input should error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only should error",
			input:   "   \n\t  ",
			wantErr: true,
		},
		{
			name:    "invalid json should error",
			input:   "not json",
			wantErr: true,
		},
		{
			name:     "valid response",
			input:    `{"response": "Hello, World!", "stats": null}`,
			wantErr:  false,
			wantResp: "Hello, World!",
		},
		{
			name: "response with stats",
			input: `{
				"response": "The capital is Paris.",
				"stats": {
					"models": {
						"gemini-2.5-pro": {
							"api": {"totalRequests": 1, "totalErrors": 0, "totalLatencyMs": 1000},
							"tokens": {"prompt": 100, "candidates": 50, "total": 150, "cached": 0, "thoughts": 10, "tool": 0}
						}
					},
					"tools": {"totalCalls": 0, "totalSuccess": 0, "totalFail": 0, "totalDurationMs": 0},
					"files": {"totalLinesAdded": 0, "totalLinesRemoved": 0}
				}
			}`,
			wantErr:  false,
			wantResp: "The capital is Paris.",
		},
		{
			name: "response with error",
			input: `{
				"response": "",
				"error": {
					"type": "AuthError",
					"message": "Invalid API key",
					"code": 401
				}
			}`,
			wantErr:  false,
			wantResp: "",
		},
	}

	parser := NewOutputParser(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := parser.ParseJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp.Response != tt.wantResp {
				t.Errorf("ParseJSON() response = %q, want %q", resp.Response, tt.wantResp)
			}
		})
	}
}

func TestParseStreamJSON(t *testing.T) {
	input := `{"type":"init","timestamp":"2025-01-01T00:00:00Z","session_id":"abc123","model":"gemini-2.5-pro"}
{"type":"message","role":"user","content":"Hello","timestamp":"2025-01-01T00:00:01Z"}
{"type":"message","role":"assistant","content":"Hi there!","timestamp":"2025-01-01T00:00:02Z"}
{"type":"result","status":"success","stats":{"models":{},"tools":{"totalCalls":0},"files":{}},"timestamp":"2025-01-01T00:00:03Z"}`

	parser := NewOutputParser(false)
	events, resp, err := parser.ParseStreamJSON(input)

	if err != nil {
		t.Errorf("ParseStreamJSON() unexpected error: %v", err)
	}

	if len(events) != 4 {
		t.Errorf("ParseStreamJSON() got %d events, want 4", len(events))
	}

	// Check event types
	expectedTypes := []string{"init", "message", "message", "result"}
	for i, event := range events {
		if event.Type != expectedTypes[i] {
			t.Errorf("Event %d type = %q, want %q", i, event.Type, expectedTypes[i])
		}
	}

	// Check init event
	if events[0].SessionID != "abc123" {
		t.Errorf("Init event session_id = %q, want %q", events[0].SessionID, "abc123")
	}
	if events[0].Model != "gemini-2.5-pro" {
		t.Errorf("Init event model = %q, want %q", events[0].Model, "gemini-2.5-pro")
	}

	// Check that result was captured
	if resp == nil {
		t.Error("ParseStreamJSON() response is nil, expected stats")
	}
}

func TestParseText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "text with whitespace",
			input:    "  \n  Hello  \n  ",
			expected: "Hello",
		},
		{
			name:     "multiline text",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
	}

	parser := NewOutputParser(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := parser.ParseText(tt.input)
			if resp.Response != tt.expected {
				t.Errorf("ParseText() = %q, want %q", resp.Response, tt.expected)
			}
		})
	}
}

func TestStatsFormatting(t *testing.T) {
	stats := &CLIStats{
		Models: map[string]ModelStats{
			"gemini-2.5-pro": {
				API:    APIStats{TotalRequests: 2, TotalErrors: 0, TotalLatencyMs: 5000},
				Tokens: TokenStats{Prompt: 1000, Candidates: 500, Total: 1500, Cached: 200, Thoughts: 100, Tool: 50},
			},
		},
		Tools: ToolStats{
			TotalCalls:      3,
			TotalSuccess:    3,
			TotalFail:       0,
			TotalDurationMs: 2000,
			ByName: map[string]ToolDetail{
				"Bash": {Count: 2, Success: 2, DurationMs: 1500},
				"Read": {Count: 1, Success: 1, DurationMs: 500},
			},
		},
		Files: FileStats{
			TotalLinesAdded:   10,
			TotalLinesRemoved: 5,
		},
	}

	formatted := FormatStats(stats)

	// Check that key elements are present
	if !strings.Contains(formatted, "gemini-2.5-pro") {
		t.Error("Formatted stats should contain model name")
	}
	if !strings.Contains(formatted, "1000") {
		t.Error("Formatted stats should contain prompt tokens")
	}
	if !strings.Contains(formatted, "Bash") {
		t.Error("Formatted stats should contain tool names")
	}
}

func TestFormatStatsSimple(t *testing.T) {
	stats := &CLIStats{
		Models: map[string]ModelStats{
			"gemini-2.5-flash": {
				Tokens: TokenStats{Prompt: 100, Candidates: 50, Total: 150},
			},
		},
		Tools: ToolStats{TotalCalls: 2},
	}

	result := FormatStatsSimple(stats)

	if !strings.Contains(result, "Tokens: 150") {
		t.Errorf("FormatStatsSimple() should contain total tokens, got: %s", result)
	}
	if !strings.Contains(result, "Tools: 2") {
		t.Errorf("FormatStatsSimple() should contain tool calls, got: %s", result)
	}
	if !strings.Contains(result, "Cost:") {
		t.Errorf("FormatStatsSimple() should contain cost, got: %s", result)
	}
}

func TestFormatStatsNil(t *testing.T) {
	result := FormatStats(nil)
	if result != "" {
		t.Errorf("FormatStats(nil) should return empty string, got: %q", result)
	}

	result = FormatStatsSimple(nil)
	if result != "No stats available" {
		t.Errorf("FormatStatsSimple(nil) should return 'No stats available', got: %q", result)
	}
}
