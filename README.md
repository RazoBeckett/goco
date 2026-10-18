# GoCo

**Go Conventional**: generate Conventional Commit messages from your git changes with an AI provider, then commit them. GoCo uses [Fang](https://charm.land) for polished help, errors, shell completions, and manpages.

## Features

- **Nine providers**: Groq (default), OpenAI, Anthropic, Google, DeepSeek, OpenRouter, xAI, Mistral, and Cerebras. Switch with one flag.
- **One prompt path**: a single key-based setup. Export an API key and go.
- **Fang CLI UX**: styled help and errors, built-in `--version`, shell completions, and manpage generation.
- **Secure input**: a password-masked prompt when a key is missing.
- **Review before commit**: inspect the message, edit it, or skip the prompt with `-y`.
- **Custom instructions**: steer the message with `-c`.

## Installation

### From source

```bash
git clone https://github.com/RazoBeckett/goco.git
cd goco
go build -o goco .
sudo mv goco /usr/local/bin/  # optional
```

Go 1.26.7 or newer is required.

### With go install

```bash
go install github.com/RazoBeckett/goco@latest
```

## Quick start

1. Export an API key for your provider. Groq is the default:

   ```bash
   export GROQ_API_KEY="your-api-key-here"
   ```

2. Change some files in a git repository. To include only staged changes, use `--staged`:

   ```bash
   cd your-project
   git add .
   goco generate --staged
   ```

3. Generate and commit:

   ```bash
   goco generate
   ```

## Usage

```bash
# Use the default provider
goco generate

# Pick a provider and model
goco generate --provider openai --model gpt-4o
goco generate -p anthropic -m claude-sonnet-4-20250514

# Work on staged changes only
goco generate --staged

# Show git status and diff first
goco generate --verbose

# Add instructions for the model
goco generate --custom-instructions "focus on the API changes"

# Edit before committing, then confirm
goco generate --edit

# Skip the confirmation prompt
goco generate --yes

# Commit on a new branch
goco generate --branch feature/new-thing

# Chain short flags (verbose, staged, yes)
goco generate -Vsy
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--provider` | `-p` | Provider to use. Defaults to `groq`. |
| `--model` | `-m` | Model to use. Defaults to the provider's recommended model. |
| `--api-key` | `-k` | API key. Prefer an environment variable. |
| `--staged` | `-s` | Use the staged diff instead of the working-tree diff. |
| `--verbose` | `-V` | Print git status and diff before generating. |
| `--yes` | `-y` | Skip the confirmation prompt. |
| `--custom-instructions` | `-c` | Extra instructions for the model. |
| `--edit` | `-e` | Open the message in `$EDITOR` before committing. |
| `--branch` | `-B` | Create and switch to a branch before committing. |

## Providers

GoCo drives providers through the [GoAI SDK](https://goai.sh). Every provider below needs only an API key.

| Provider | `--provider` | Default model | Key environment variable |
|----------|--------------|---------------|--------------------------|
| Groq | `groq` | `llama-3.3-70b-versatile` | `GROQ_API_KEY` |
| OpenAI | `openai` | `gpt-4o` | `OPENAI_API_KEY` |
| Anthropic | `anthropic` | `claude-sonnet-4-20250514` | `ANTHROPIC_API_KEY` |
| Google | `google` | `gemini-2.5-flash` | `GOOGLE_GENERATIVE_AI_API_KEY` |
| DeepSeek | `deepseek` | `deepseek-chat` | `DEEPSEEK_API_KEY` |
| OpenRouter | `openrouter` | `openai/gpt-4o` | `OPENROUTER_API_KEY` |
| xAI | `xai` | `grok-3` | `XAI_API_KEY` |
| Mistral | `mistral` | `mistral-large-latest` | `MISTRAL_API_KEY` |
| Cerebras | `cerebras` | `llama-3.3-70b` | `CEREBRAS_API_KEY` |

For any provider, GoCo also accepts `GOCO_<PROVIDER>_KEY`. For example, `GOCO_GROQ_KEY` works as a fallback for Groq. This keeps keys out of shell history when you prefer not to use the provider's own variable.

## Configuration

The config file lives at `~/.config/goco/config.toml`, following the XDG Base Directory specification. The file is optional. If it is missing, GoCo uses Groq and the standard environment variables.

```toml
# ~/.config/goco/config.toml

[General]
default_provider = "groq"

# Optional per-provider overrides.
[providers.groq]
api_key_env_variable = "GROQ_API_KEY"
model = "llama-3.3-70b-versatile"
```

- `default_provider` sets the provider used when `--provider` is absent.
- `providers.<name>.api_key_env_variable` overrides the environment variable name.
- `providers.<name>.model` sets a default model for that provider.

### Environment variables

| Variable | Description |
|----------|-------------|
| `GROQ_API_KEY` | Groq API key. |
| `OPENAI_API_KEY` | OpenAI API key. |
| `ANTHROPIC_API_KEY` | Anthropic API key. |
| `GOOGLE_GENERATIVE_AI_API_KEY` | Google API key. |
| `DEEPSEEK_API_KEY` | DeepSeek API key. |
| `OPENROUTER_API_KEY` | OpenRouter API key. |
| `XAI_API_KEY` | xAI API key. |
| `MISTRAL_API_KEY` | Mistral API key. |
| `CEREBRAS_API_KEY` | Cerebras API key. |
| `GOCO_<PROVIDER>_KEY` | Fallback for any provider. |
| `XDG_CONFIG_HOME` | Base directory for the config file. Defaults to `~/.config`. |

## Example output

```
┌─────────────────────────────────────────────────────────────┐
│ feat(auth): add OAuth2 login with Google                    │
│                                                             │
│ - Add the OAuth2 flow for Google sign-in                    │
│ - Store the session after the callback                      │
└─────────────────────────────────────────────────────────────┘

Proceed with this commit?   Yes   No
```

Verbose mode adds two boxes before the message: git status and git diff. The theme is orange.

## Conventional Commits

GoCo writes messages in the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Supported types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`. Breaking changes use `!` before the colon, for example `feat!: drop the v1 API`.

The generated subject must be 72 characters or fewer and must match the format. If it does not, GoCo stops and asks you to edit it.

## Development

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .
```

The CI workflow runs build, vet, formatting, and tests on every pull request.

## Tech stack

- Go 1.26+
- [Fang](https://charm.land) and [Cobra](https://github.com/spf13/cobra) for the CLI
- [Charm v2](https://charm.land/blog/v2/) for the terminal UI (Bubble Tea, Bubbles, Lip Gloss)
- [GoAI SDK](https://goai.sh) for provider access
- [BurntSushi/toml](https://github.com/BurntSushi/toml) for config

## License

MIT. See [LICENSE](LICENSE).
