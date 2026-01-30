# Agent Guidelines for GoCo

## Project Overview
CLI AI assistant (GoCo - Go Conventional) that generates Conventional Commit messages using Google Gemini AI. Built with Cobra for CLI framework and Charm Bracelet libraries for TUI components.

## Build & Test Commands
- **Build**: `go build -o goco .`
- **Test all**: `go test ./...`
- **Test single package**: `go test ./config` (or any package path)
- **Lint**: `go vet ./...`
- **Format**: `go fmt ./...`
- **Tidy dependencies**: `go mod tidy`

## Code Style
- **Go version**: 1.24.6+
- **Imports**: Standard library first, then external packages, then local packages (see cmd/generate.go:3-19)
- **Formatting**: Use `go fmt` - tabs for indentation, standard Go formatting
- **Naming**: camelCase for private, PascalCase for exported. Acronyms all caps (e.g., `apiKey`, `GetGeminiApiKey`)
- **Error handling**: Always check errors immediately after function calls, use `log.Fatalf()` for fatal errors
- **Comments**: Export comments on all exported functions/types. Use `//` for line comments
- **Types**: Define types for configs and models (see config/config.go). Use struct tags for TOML/JSON

## Project Structure
- `cmd/` - Cobra command definitions
- `config/` - Configuration loading and management
- Follow XDG Base Directory spec for config files (~/.config/goco/)
