# GoCo 🚀

**Go Conventional** - AI-powered conventional commit message generator with a beautiful Terminal User Interface.

GoCo transforms your git workflow by automatically generating meaningful conventional commit messages using AI providers (Google Gemini or Groq), all wrapped in a gorgeous TUI built with Charm Bracelet libraries.

![GOCO_PREVIEW](demo.gif)

## ✨ Features

- 🤖 **Multi-Provider AI**: Choose between Google Gemini or Groq (Llama models) for commit message generation
- 🎨 **Beautiful TUI**: Modern terminal interface with styled output, loading spinners, and interactive prompts
- 🔒 **Secure Input**: Password-masked API key prompts when credentials are missing
- ⚙️ **Smart Config**: TOML-based configuration with XDG Base Directory support and multi-provider support
- 👁️ **Verbose Mode**: Optional detailed view of git status and diff in styled boxes
- 🎯 **Zero Setup**: Works out of the box with minimal configuration
- 🔧 **Custom Instructions**: Add custom instructions to tailor commit messages to your needs
- 📋 **Model Selection**: List and select from available AI models for each provider

## 📦 Installation

### From Source

```bash
git clone https://github.com/RazoBeckett/goco.git
cd goco
go build -o goco .
sudo mv goco /usr/local/bin/  # Optional: install globally
```

### Using Go Install

```bash
go install github.com/RazoBeckett/goco@latest
```

## 🚀 Quick Start

1. **Choose your AI provider and set up your API key**:

   **Google Gemini** (get one from [Google AI Studio](https://aistudio.google.com/apikey)):
   ```bash
   export GOCO_GEMINI_KEY="your-api-key-here"
   ```

   **Groq** (get one from [Groq Console](https://console.groq.com)):
   ```bash
   export GOCO_GROQ_KEY="your-api-key-here"
   ```

2. **Navigate to your git repository** and stage your changes:
   ```bash
   cd your-project
   git add .
   ```

3. **Generate a commit message**:
   ```bash
   # Use default provider (Gemini)
   goco generate

   # Use Groq provider
   goco generate --provider groq

   # With custom instructions
   goco generate --custom-instructions "focus on the backend changes"

   # With specific model
   goco generate --provider gemini --model gemini-2.5-flash
   ```

4. **List available models**:
   ```bash
   # List models for default provider
   goco models

   # List models for specific provider
   goco models --provider groq
   ```

That's it! GoCo will analyze your staged changes and generate a beautiful conventional commit message, then automatically commit it.

## 💡 Usage

### Basic Usage

```bash
# Generate commit message for staged changes
goco generate

# Show detailed git status and diff
goco generate --verbose

# Use short flag for verbose mode
goco generate -v

# Use Groq provider instead of Gemini
goco generate --provider groq

# Use specific model
goco generate --provider groq --model llama-3.3-70b-versatile

# Add custom instructions for the AI
goco generate --custom-instructions "make the message concise"

# Use staged diff instead of working directory
goco generate --staged
```

### Listing Available Models

```bash
# List models for default provider
goco models

# List models for a specific provider
goco models --provider gemini
goco models --provider groq
```

### Interactive Features

- **Missing API Key**: GoCo will prompt you securely with a password-masked input
- **Loading Indicator**: Beautiful animated spinner while generating messages
- **Styled Output**: Commit messages appear in elegant green-bordered boxes
- **Git Info**: Verbose mode shows git status and diff in separate styled containers
- **Auto-Commit**: After generating, GoCo automatically stages and commits your changes

## ⚙️ Configuration

GoCo uses a TOML configuration file located at `~/.config/goco/config.toml` (following XDG Base Directory specification).

### Default Configuration

```toml
# ~/.config/goco/config.toml
api_key_gemini_env_variable = "GOCO_GEMINI_KEY"
api_key_groq_env_variable = "GOCO_GROQ_KEY"
default_provider = "gemini"
```

Note: The config file follows a [General] TOML table in the application configuration.
When using a full config file, prefer the following structure under a [General]
header to avoid ambiguity with top-level keys:

```toml
[General]
api_key_gemini_env_variable = "GOCO_GEMINI_KEY"
api_key_groq_env_variable = "GOCO_GROQ_KEY"
default_provider = "gemini"
```

This README snippet documents the configuration format used by config.LoadConfig and
helps users migrate from older top-level keys to the new [General] table if they
previously placed these keys at the top level of the file.

### Custom Environment Variables

You can customize which environment variable GoCo looks for for each provider:

```toml
# Use different environment variable names
api_key_gemini_env_variable = "MY_CUSTOM_GEMINI_KEY"
api_key_groq_env_variable = "MY_CUSTOM_GROQ_KEY"
```

Then set your custom variables:
```bash
export MY_CUSTOM_GEMINI_KEY="your-api-key-here"
export MY_CUSTOM_GROQ_KEY="your-api-key-here"
```

### Setting Default Provider

You can set a default AI provider in your configuration:

```toml
# Set Groq as the default provider
default_provider = "groq"
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GOCO_GEMINI_KEY` | - | Your Google Gemini API key |
| `GOCO_GROQ_KEY` | - | Your Groq API key |
| `XDG_CONFIG_HOME` | `~/.config` | Base directory for config files |

## 🎨 Example Output

### Standard Mode
```
┌─────────────────────────────────────────────────────────────┐
│ feat(auth): implement OAuth2 authentication with Google     │
│                                                             │
│ - Add OAuth2 flow for Google authentication                │
│ - Integrate user session management                        │
│ - Add proper error handling for auth failures              │
└─────────────────────────────────────────────────────────────┘
```

### Verbose Mode
Shows additional styled boxes with:
- 📊 **Git Status**: Current repository status in a blue-bordered box
- 📝 **Git Diff**: Detailed changes in a yellow-bordered box
- ✅ **Commit Message**: Generated message in a green-bordered box

## 🔧 Development

### Building

```bash
go build -o goco .
```

### Running Tests

```bash
go test ./...
```

### Linting & Formatting

```bash
go vet ./...
go fmt ./...
```

### Clean Dependencies

```bash
go mod tidy
```

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **TUI Components**: [Charm Bracelet](https://charm.sh/)
  - [Huh](https://github.com/charmbracelet/huh) - Interactive forms
  - [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
  - [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- **AI Providers**: 
  - [Google Gemini API](https://ai.google.dev/)
  - [Groq API](https://console.groq.com/) (Llama models)
- **Config**: [Viper](https://github.com/spf13/viper) with TOML

## 📝 Conventional Commits

GoCo generates commits following the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Supported Types
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/amazing-feature`
3. Commit your changes: `git commit -m "feat: add amazing feature"`
4. Push to the branch: `git push origin feat/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Conventional Commits](https://www.conventionalcommits.org/) for the commit message standard
- [Charm Bracelet](https://charm.sh/) for the amazing TUI libraries
- [Google](https://ai.google.dev/) for the Gemini AI API
- [Groq](https://console.groq.com/) for the fast Llama models
- The Go community for excellent tooling and libraries

---

Made with ❤️ and Go. Transform your git workflow today!
