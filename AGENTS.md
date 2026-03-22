# Agent Info

Generally speaking, you should browse the codebase to figure out what is going on.

We have a few "philosophies" I want to make sure we honor throughout development:

### 1. Performance above all else

When in doubt, do the thing that makes the app feel the fastest to use.

This includes things like

- Optimistic updates
- Avoiding waterfalls in anything from js to file fetching

### 2. Good defaults

Users should expect things to behave well by default. Less config is best.

### 3. Security

We want to make things convenient, but we don't want to be insecure. Be thoughtful about how things are implemented.

## Project Overview
GoCo is a Fang-based CLI that generates and applies Conventional Commit messages from git changes using Gemini or Groq. The command surface follows Fang conventions for help, errors, `--version`, shell completions, and manpage generation, while command logic is split into focused internal packages for CLI wiring, provider integrations, git operations, and config loading.

## Code Style
- **Go version**: 1.25.5+
- Prefer Fang/Cobra command constructors with explicit dependencies over package-level mutable globals.
- Keep business logic out of command bootstrap files; `internal/cli` should orchestrate, while `internal/git`, `internal/ai`, and `internal/config` own their domains.
- Return errors up to Fang instead of calling `log.Fatal` or `os.Exit` from internal packages.
- Preserve fast-feeling CLI behavior: avoid unnecessary subprocesses, keep prompts purposeful, and default to single-pass flows.

## Project Structure
- `main.go`: Fang entrypoint that executes the root Cobra command with Fang-provided UX.
- `internal/cli/`: Root command, subcommands, flag binding, prompts, spinners, and styled command output.
- `internal/ai/`: Gemini and Groq provider implementations, provider factory, and prompt construction.
- `internal/git/`: Git repository helpers for status, diff, staging, branch creation, and commit execution.
- `internal/config/`: XDG-aware config loading and provider/env-var defaults.
