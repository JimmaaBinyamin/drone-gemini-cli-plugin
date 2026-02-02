# Contributing to drone-gemini-cli-plugin

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

```bash
# Clone the repository
git clone https://github.com/JimmaaBinyamin/drone-gemini-cli-plugin.git
cd drone-gemini-cli-plugin

# Install dependencies
go mod download

# Build the plugin
go build -o drone-gemini-cli-plugin .

# Run tests
go test -v ./...
```

## How to Contribute

### Reporting Issues

- Search existing issues before creating a new one
- Provide detailed reproduction steps
- Include Go version and OS information

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Run tests: `go test -v ./...`
5. Commit with clear messages
6. Push and create a Pull Request

### Code Style

- Follow standard Go formatting (`go fmt`)
- Add tests for new functionality
- Update documentation as needed

## Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Local plugin test (requires API key)
PLUGIN_PROMPT="Hello" \
PLUGIN_API_KEY="your-key" \
./drone-gemini-cli-plugin
```

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
