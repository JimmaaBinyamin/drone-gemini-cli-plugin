# drone-gemini-cli-plugin

English | [中文](README_zh.md)

A Drone CI/CD plugin that leverages Gemini CLI's Headless mode for AI-powered code analysis, review, and automated fixes.

## Features

- **AI Code Review**: Analyze code for security issues, bugs, and improvements
- **File-Based Prompts**: Load review guidelines from markdown files (`prompt_file`)
- **Context Files**: Provide additional architecture/docs context (`context_file`)
- **YOLO Mode**: Auto-approve all AI actions for automated code modifications
- **Tool Execution**: AI agent can read files, run shell commands, and more
- **Git Integration**: Automatically include commit diff for targeted reviews
- **Detailed Statistics**: Token usage, costs, and tool call metrics
- **Dual Auth Support**: Gemini API Key or Vertex AI Service Account

## Quick Start

### Option A: Vertex AI with Service Account (Recommended)

For enterprise use or when running on AWS/Azure/on-prem infrastructure.

```bash
# 1. Create GCP Service Account with Vertex AI User role
gcloud iam service-accounts create gemini-drone-sa
gcloud projects add-iam-policy-binding YOUR_PROJECT \
  --member="serviceAccount:gemini-drone-sa@YOUR_PROJECT.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# 2. Download key and add to Drone
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=gemini-drone-sa@YOUR_PROJECT.iam.gserviceaccount.com

drone secret add --repository your-org/your-repo \
  --name gcp_credentials --data @sa-key.json
```

```yaml
kind: pipeline
type: docker
name: ai-review

steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "Review this code for security issues and bugs"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### Option B: Gemini API Key (Simplest)

Get a free API key from [Google AI Studio](https://aistudio.google.com/apikey).

```bash
drone secret add --repository your-org/your-repo \
  --name gemini_api_key --data "AIzaSy..."
```

```yaml
steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "Review this code for bugs and security issues"
      model: gemini-3-flash-preview
      api_key:
        from_secret: gemini_api_key
```

## Configuration

| Parameter | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `prompt` | `PLUGIN_PROMPT` | string | required* | AI instruction/prompt (*not required if `prompt_file` is set) |
| `prompt_file` | `PLUGIN_PROMPT_FILE` | string | | File path to read prompt from (overrides `prompt`) |
| `context_file` | `PLUGIN_CONTEXT_FILE` | string | | File path to read additional context (passed via stdin) |
| `target` | `PLUGIN_TARGET` | string | `.` | Working directory (usually `/drone/src`) |
| `model` | `PLUGIN_MODEL` | string | `gemini-2.5-pro` | Model to use (recommended: `gemini-3-flash-preview`) |
| `output_format` | `PLUGIN_OUTPUT_FORMAT` | string | `json` | `text`, `json`, `stream-json` |
| `yolo` | `PLUGIN_YOLO` | bool | `false` | Auto-approve all actions (enables file modifications) |
| `approval_mode` | `PLUGIN_APPROVAL_MODE` | string | | Override approval mode |
| `include_dirs` | `PLUGIN_INCLUDE_DIRS` | string | | Comma-separated directories to include |
| `stdin_input` | `PLUGIN_STDIN_INPUT` | string | | Additional content passed via stdin |
| `api_key` | `PLUGIN_API_KEY` | string | | Gemini API Key (Google AI Studio) |
| `gcp_project` | `PLUGIN_GCP_PROJECT` | string | | GCP Project ID (Vertex AI) |
| `gcp_location` | `PLUGIN_GCP_LOCATION` | string | `us-central1` | GCP Location (use `global` for gemini-3-*) |
| `gcp_credentials` | `PLUGIN_GCP_CREDENTIALS` | string | | Service Account JSON content |
| `git_diff` | `PLUGIN_GIT_DIFF` | bool | `false` | Include git commit diff in context |
| `timeout` | `PLUGIN_TIMEOUT` | int | `300` | Timeout in seconds |
| `debug` | `PLUGIN_DEBUG` | bool | `false` | Enable debug output |

## Examples

### 1. Basic Code Review (Inline Prompt + Git Diff)

```yaml
steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: |
        Review this code for:
        - Security vulnerabilities
        - Potential bugs
        - Code quality issues
        Give a brief summary.
      git_diff: true
      model: gemini-3-flash-preview
      output_format: json
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      event: pull_request
```

### 2. File-Based Review Guidelines (prompt_file)

Store your review standards in a markdown file and the plugin will load it as the prompt:

```yaml
steps:
  - name: review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt_file: ".review-guidelines.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

Example `.review-guidelines.md`:

```markdown
# Code Review Guidelines

You are a senior security engineer. Check for:
- SQL Injection: parameterized queries?
- XSS: user input properly escaped?
- Sensitive Data Exposure: passwords/tokens in responses?

Output format: CRITICAL / WARNING / INFO
```

### 3. Architecture Context Review (context_file)

Provide additional architecture documentation as context for the AI:

```yaml
steps:
  - name: review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "Review this code against our architecture guidelines"
      context_file: "docs/architecture.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### 4. Full Review (prompt_file + context_file)

Combine review guidelines with architecture context for the most comprehensive review:

```yaml
steps:
  - name: full-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt_file: ".review-guidelines.md"
      context_file: "docs/architecture.md"
      git_diff: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### 5. YOLO Mode - AI Auto-Fix

> **Warning**: YOLO mode allows AI to modify files and execute commands automatically. Use on non-protected branches only.

```yaml
steps:
  - name: auto-fix
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: |
        Fix the SQL injection vulnerabilities by using parameterized queries.
        Fix the XSS vulnerabilities by escaping user input.
        Only modify source files, do not create new files.
        Show me the diff of your changes.
      yolo: true
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      branch: fix/*
```

### 6. Generate Release Notes

```yaml
steps:
  - name: release-notes
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin:v0.1.5
    settings:
      prompt: "Generate release notes from recent commits in CHANGELOG format"
      git_diff: true
      output_format: text
      model: gemini-3-flash-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
    when:
      event: tag
```

## Local Testing

```bash
# Build the plugin
go build -o drone-gemini-cli-plugin .

# Test with Vertex AI
PLUGIN_PROMPT="Describe this project" \
PLUGIN_GCP_PROJECT="your-project-id" \
PLUGIN_GCP_LOCATION="global" \
PLUGIN_GCP_CREDENTIALS="$(cat service-account.json)" \
PLUGIN_MODEL="gemini-3-flash-preview" \
./drone-gemini-cli-plugin

# Test with Gemini API Key
PLUGIN_PROMPT="Describe this project" \
PLUGIN_API_KEY="your-api-key" \
PLUGIN_MODEL="gemini-3-flash-preview" \
./drone-gemini-cli-plugin

# Test YOLO mode (allows file modifications)
PLUGIN_PROMPT="Fix security issues in main.go" \
PLUGIN_GCP_PROJECT="your-project-id" \
PLUGIN_GCP_LOCATION="global" \
PLUGIN_GCP_CREDENTIALS="$(cat service-account.json)" \
PLUGIN_YOLO="true" \
./drone-gemini-cli-plugin
```

## Building Docker Image

```bash
docker build -t ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin .
docker push ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
```

## Drone Plugin Environment Variables

The plugin automatically receives these Drone environment variables:

| Variable | Description |
|----------|-------------|
| `DRONE_COMMIT_SHA` | Current commit SHA (used for git_diff) |
| `DRONE_REPO_NAME` | Repository name |
| `DRONE_BUILD_EVENT` | Build event type (push, pull_request, tag) |

## License

Apache License 2.0
