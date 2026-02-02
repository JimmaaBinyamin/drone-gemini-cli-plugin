package plugin

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "empty prompt should fail",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "valid config with prompt",
			config: Config{
				Prompt: "test prompt",
			},
			wantErr: false,
		},
		{
			name: "full config should pass",
			config: Config{
				Prompt:       "Review this code",
				Target:       ".",
				Model:        "gemini-2.5-pro",
				OutputFormat: "json",
				Timeout:      300,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectAuthMode(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected AuthMode
	}{
		{
			name:     "no auth configured",
			config:   Config{Prompt: "test"},
			expected: AuthModeNone,
		},
		{
			name: "gemini api key only",
			config: Config{
				Prompt: "test",
				APIKey: "test-api-key",
			},
			expected: AuthModeGeminiAPIKey,
		},
		{
			name: "vertex ai with api key",
			config: Config{
				Prompt:     "test",
				APIKey:     "test-api-key",
				GCPProject: "my-project",
			},
			expected: AuthModeVertexAI,
		},
		{
			name: "vertex ai with service account",
			config: Config{
				Prompt:         "test",
				GCPProject:     "my-project",
				GCPCredentials: `{"type":"service_account"}`,
			},
			expected: AuthModeVertexAI,
		},
		{
			name: "credentials without project should be none",
			config: Config{
				Prompt:         "test",
				GCPCredentials: `{"type":"service_account"}`,
			},
			expected: AuthModeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.DetectAuthMode()
			if result != tt.expected {
				t.Errorf("DetectAuthMode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
