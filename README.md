# drone-gemini-cli-plugin

English | [中文](README_zh.md)

A Drone CI/CD plugin that leverages Gemini CLI's Headless mode for AI-powered code analysis, review, and automated fixes.

## Features

- **AI Code Review**: Analyze code for security issues, bugs, and improvements
- **YOLO Mode**: Auto-approve all AI actions for automated code modifications
- **Tool Execution**: Supports Bash, file read/write, web search via Gemini CLI
- **Git Integration**: Analyze specific commits or PR diffs
- **Detailed Statistics**: Token usage, costs, and tool call metrics
- **Dual Auth Support**: Gemini API Key or Vertex AI Service Account

## Quick Start

### Option A: Gemini API Key (Simplest)

Get a free API key from [Google AI Studio](https://aistudio.google.com/apikey).

```bash
# Add secret to Drone
drone secret add --repository your-org/your-repo \
  --name gemini_api_key --data "AIzaSy..."
```

```yaml
kind: pipeline
type: docker
name: ai-review

steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "Review this code for bugs and security issues"
      model: gemini-2.5-flash
      api_key:
        from_secret: gemini_api_key
```

### Option B: Vertex AI with Service Account

For enterprise use or when running on AWS/Azure/on-prem infrastructure.

```bash
# 1. Create GCP Service Account with Vertex AI User role
gcloud iam service-accounts create gemini-cli-sa
gcloud projects add-iam-policy-binding YOUR_PROJECT \
  --member="serviceAccount:gemini-cli-sa@YOUR_PROJECT.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# 2. Download key and add to Drone
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=gemini-cli-sa@YOUR_PROJECT.iam.gserviceaccount.com

drone secret add --repository your-org/your-repo \
  --name gcp_credentials --data @sa-key.json
```

```yaml
kind: pipeline
type: docker
name: ai-review

steps:
  - name: code-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "Review this code for security issues"
      model: gemini-3-pro-preview
      gcp_project: your-gcp-project-id
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

## Configuration

| Parameter | Environment Variable | Type | Default | Description |
|-----------|---------------------|------|---------|-------------|
| `prompt` | `PLUGIN_PROMPT` | string | **required** | AI instruction/prompt |
| `target` | `PLUGIN_TARGET` | string | `.` | Working directory (usually `/drone/src`) |
| `model` | `PLUGIN_MODEL` | string | `gemini-2.5-pro` | Model to use |
| `output_format` | `PLUGIN_OUTPUT_FORMAT` | string | `json` | `text`, `json`, `stream-json` |
| `yolo` | `PLUGIN_YOLO` | bool | `false` | Auto-approve all actions (enables file modifications) |
| `approval_mode` | `PLUGIN_APPROVAL_MODE` | string | | Override approval mode |
| `include_dirs` | `PLUGIN_INCLUDE_DIRS` | string | | Comma-separated directories to include |
| `api_key` | `PLUGIN_API_KEY` | string | | Gemini API Key (Google AI Studio) |
| `gcp_project` | `PLUGIN_GCP_PROJECT` | string | | GCP Project ID (Vertex AI) |
| `gcp_location` | `PLUGIN_GCP_LOCATION` | string | `us-central1` | GCP Location (use `global` for gemini-3-*) |
| `gcp_credentials` | `PLUGIN_GCP_CREDENTIALS` | string | | Service Account JSON content |
| `git_diff` | `PLUGIN_GIT_DIFF` | bool | `false` | Include git commit diff in context |
| `timeout` | `PLUGIN_TIMEOUT` | int | `300` | Timeout in seconds |
| `debug` | `PLUGIN_DEBUG` | bool | `false` | Enable debug output |

## Examples

### PR Code Review

```yaml
steps:
  - name: ai-review
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: |
        Review this PR for:
        - Security vulnerabilities
        - Performance issues
        - Code quality and maintainability
      git_diff: true
      model: gemini-2.5-flash
      api_key:
        from_secret: gemini_api_key
    when:
      event: pull_request
```

### Auto-Fix with YOLO Mode

**Warning**: YOLO mode allows AI to modify files automatically. Use with caution.

```yaml
steps:
  - name: auto-fix
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "Fix all linting errors and format the code"
      yolo: true
      api_key:
        from_secret: gemini_api_key
    when:
      branch: fix/*
```

### Security Audit with Vertex AI

```yaml
steps:
  - name: security-audit
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: |
        Perform comprehensive security audit:
        1. Check for SQL/NoSQL injection
        2. Review authentication and authorization
        3. Identify sensitive data exposure
        4. Check for dependency vulnerabilities
      include_dirs: "src,api,lib"
      model: gemini-3-pro-preview
      gcp_project: my-project
      gcp_location: global
      gcp_credentials:
        from_secret: gcp_credentials
```

### Generate Release Notes

```yaml
steps:
  - name: release-notes
    image: ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin
    settings:
      prompt: "Generate release notes from the recent commits in CHANGELOG format"
      git_diff: true
      api_key:
        from_secret: gemini_api_key
    when:
      event: tag
```

## Local Testing

```bash
# Build the plugin
go build -o drone-gemini-cli-plugin .

# Test with Gemini API Key
PLUGIN_PROMPT="Describe this project" \
PLUGIN_API_KEY="your-api-key" \
PLUGIN_MODEL="gemini-2.5-flash" \
./drone-gemini-cli-plugin

# Test with Vertex AI (using JSON file)
PLUGIN_PROMPT="Describe this project" \
PLUGIN_GCP_PROJECT="your-project-id" \
PLUGIN_GCP_LOCATION="global" \
PLUGIN_GCP_CREDENTIALS="$(cat service-account.json)" \
PLUGIN_MODEL="gemini-3-pro-preview" \
./drone-gemini-cli-plugin

# Test YOLO mode (allows file modifications)
PLUGIN_PROMPT="Create a hello.txt file with greeting" \
PLUGIN_API_KEY="your-api-key" \
PLUGIN_YOLO="true" \
./drone-gemini-cli-plugin
```

## Building Docker Image

```bash
# Build the image
docker build -t ghcr.io/jimmaabinyamin/drone-gemini-cli-plugin .

# Or with custom registry
docker build -t your-registry.com/drone-gemini-cli:latest .
docker push your-registry.com/drone-gemini-cli:latest
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
