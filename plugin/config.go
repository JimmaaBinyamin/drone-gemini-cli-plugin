package plugin

// Config holds the plugin configuration from environment variables.
// Drone CI injects these as PLUGIN_* environment variables.
type Config struct {
	// Prompt is the instruction for the AI (required)
	Prompt string `envconfig:"PROMPT" required:"true"`

	// Target is the working directory for gemini CLI (optional, defaults to ".")
	Target string `envconfig:"TARGET" default:"."`

	// Model specifies which AI model to use
	Model string `envconfig:"MODEL" default:"gemini-2.5-pro"`

	// OutputFormat specifies the output format: text, json, stream-json
	OutputFormat string `envconfig:"OUTPUT_FORMAT" default:"json"`

	// Yolo enables auto-approval of all actions (--yolo flag)
	Yolo bool `envconfig:"YOLO" default:"false"`

	// ApprovalMode sets the approval mode (e.g., auto_edit)
	ApprovalMode string `envconfig:"APPROVAL_MODE"`

	// IncludeDirs specifies additional directories to include (comma-separated)
	IncludeDirs string `envconfig:"INCLUDE_DIRS"`

	// Debug enables debug output
	Debug bool `envconfig:"DEBUG" default:"false"`

	// Timeout in seconds for CLI execution (default 300s = 5 minutes)
	Timeout int `envconfig:"TIMEOUT" default:"300"`

	// GitDiff enables analyzing the last commit diff
	GitDiff bool `envconfig:"GIT_DIFF" default:"false"`

	// GitCommitSHA to analyze (auto-detected from DRONE_COMMIT_SHA if empty)
	GitCommitSHA string `envconfig:"GIT_COMMIT_SHA"`

	// StdinInput is additional content to pass via stdin
	StdinInput string `envconfig:"STDIN_INPUT"`

	// --- Authentication Options ---
	// Compatible with drone-gemini-plugin configuration

	// APIKey can be used for:
	// 1. Gemini API Key (Google AI Studio) - when GCPProject is NOT set
	// 2. Google API Key (Vertex AI) - when GCPProject IS set
	// Maps to GEMINI_API_KEY or GOOGLE_API_KEY environment variable for gemini CLI
	APIKey string `envconfig:"API_KEY"`

	// GCPProject is the Google Cloud Project ID (for Vertex AI)
	// Maps to GOOGLE_CLOUD_PROJECT environment variable for gemini CLI
	GCPProject string `envconfig:"GCP_PROJECT"`

	// GCPLocation is the Google Cloud location (for Vertex AI)
	// Maps to GOOGLE_CLOUD_LOCATION environment variable for gemini CLI
	GCPLocation string `envconfig:"GCP_LOCATION" default:"us-central1"`

	// GCPCredentials is the path to or content of service account JSON (for Vertex AI)
	// Maps to GOOGLE_APPLICATION_CREDENTIALS environment variable for gemini CLI
	GCPCredentials string `envconfig:"GCP_CREDENTIALS"`
}

// AuthMode represents the authentication mode
type AuthMode int

const (
	AuthModeNone         AuthMode = iota
	AuthModeGeminiAPIKey          // Use GEMINI_API_KEY (Google AI Studio)
	AuthModeVertexAI              // Use GOOGLE_CLOUD_PROJECT + GOOGLE_API_KEY or GOOGLE_APPLICATION_CREDENTIALS
)

// DetectAuthMode automatically detects which authentication mode to use
// Compatible with drone-gemini-plugin:
// - APIKey + GCPProject = Vertex AI
// - APIKey alone = Gemini API Key
// - GCPCredentials + GCPProject = Vertex AI with service account
func (c *Config) DetectAuthMode() AuthMode {
	// Option 1: Vertex AI with API Key
	if c.APIKey != "" && c.GCPProject != "" {
		return AuthModeVertexAI
	}

	// Option 2: Vertex AI with Service Account
	if c.GCPCredentials != "" && c.GCPProject != "" {
		return AuthModeVertexAI
	}

	// Option 3: Gemini API Key (Google AI Studio)
	if c.APIKey != "" {
		return AuthModeGeminiAPIKey
	}

	return AuthModeNone
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Prompt == "" {
		return ErrPromptRequired
	}
	return nil
}
