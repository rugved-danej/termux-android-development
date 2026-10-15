# Contributing to termux-android-dev (tadev)

Thank you for your interest in contributing to `tadev`!

## Code of Conduct & Core Rules

1. **Zero Emojis**: We enforce a strict zero-emoji policy across all source code, comments, terminal outputs, spinner texts, log messages, and git commits.
2. **Clean Commit History**: The main branch maintains a linear, clean commit history.
3. **Cross-Platform Compatibility**: Changes must remain compatible across Termux Android (`arm64`, `arm`) and standard Linux distributions (`amd64`, `386`). Avoid hardcoding paths unique to Termux without proper Linux fallbacks.

## Development Setup

### Prerequisites
- Go 1.22 or newer
- Git

### Building Locally

```bash
git clone https://github.com/rugved-danej/termux-android-development.git
cd termux-android-development

# Run test suite
go test -v ./...

# Build native binary
./build.sh
```

### Running Tests

Before submitting changes, ensure all tests pass:

```bash
go test -v ./...
go vet ./...
gofmt -l .
```

## Pull Request Guidelines

1. Ensure all tests and `go vet` pass.
2. Format all Go files with `gofmt -w .`.
3. Follow the pull request template provided in `.github/pull_request_template.md`.
