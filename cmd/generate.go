package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/razobeckett/goco/providers"
	"github.com/spf13/cobra"
)

var (
	apiKey         string
	provider       string
	model          string
	commitType     string
	breakingChange bool
	// Use the correctly spelled `staged` flag only. The old `--stagged`
	// alias has been removed.
	staged             bool
	verbose            bool
	customInstructions string
	edit               bool
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			MarginBottom(1)

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			MarginTop(1)

	statusHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true).
				Background(lipgloss.Color("#065F46")).
				Padding(0, 1)

	diffHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true).
			Background(lipgloss.Color("#1E3A8A")).
			Padding(0, 1)

	statusBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1).
			MarginBottom(1).
			Width(80)

	diffBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6")).
			Padding(1).
			MarginBottom(1).
			Width(80)

	commitMessageHeaderStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).
					Background(lipgloss.Color("#059669")).
					Padding(0, 1)

	commitMessageBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#10B981")).
				Padding(1).
				MarginBottom(1).
				Width(80)
)

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case string:
		if msg == "done" {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// getStagedFiles returns the list of paths currently staged for commit in the
// repository at dir. If dir is empty the current working directory is used.
func getStagedFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	s := strings.Fields(strings.TrimSpace(string(out)))
	return s, nil
}

func newSpinnerModel(message string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	return spinnerModel{
		spinner: s,
		message: message,
	}
}

func promptForApiKey(envVar, providerName string) (string, error) {
	var apiKey string

	fmt.Println(titleStyle.Render(fmt.Sprintf("🔑 %s API Key Required", providerName)))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Please enter your %s API key:", providerName)).
				Description("Your key will be set for this session only").
				Value(&apiKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("API key cannot be empty")
					}
					return nil
				}),
		),
	)

	err := form.Run()
	if err != nil {
		return "", err
	}

	// Set for current session
	os.Setenv(envVar, apiKey)

	// Show helpful note
	note := fmt.Sprintf(`
		%s

		To avoid this prompt in the future, add this to your shell profile:
		export %s="your-api-key-here"

		For bash: ~/.bashrc or ~/.bash_profile
		For zsh: ~/.zshrc
		For fish: ~/.config/fish/config.fish`,
		noteStyle.Render("✅ API key set for this session!"), envVar)

	fmt.Println(note)
	fmt.Println()

	return apiKey, nil
}

func editCommitMessage(message string) (string, error) {
	tmpFile, err := os.CreateTemp("", "goco-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(message); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editors := []string{"vim", "nano", "vi"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return "", fmt.Errorf("failed to find editor: %w", ErrNoEditor)
	}

	editCmd := exec.Command(editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return "", fmt.Errorf("editor %q failed: %w", editor, err)
	}

	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited message: %w", err)
	}

	editedMessage := strings.TrimSpace(string(editedContent))
	if editedMessage == "" {
		return message, nil
	}

	return editedMessage, nil
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate a commit message using AI",
	Example: "  goco generate --provider gemini --verbose\n  goco generate -e --custom-instructions \"focus on api changes\"",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		validProviders := map[string]bool{"gemini": true, "groq": true}

		if provider != "" && !validProviders[provider] {
			return &ValidationError{
				Field:   "provider",
				Message: fmt.Sprintf("invalid provider %q", provider),
				Help:    "supported providers: gemini, groq. Use --provider flag or configure default in config.",
			}
		}

		if model != "" && provider == "" {
			return &ValidationError{
				Field:   "model",
				Message: "model requires provider",
				Help:    "specify --provider when using --model flag",
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Use flag value for provider if provided, otherwise get from config
		if provider == "" {
			provider = GetConfig().GetDefaultProvider()
		}

		var aiProvider providers.Provider
		var err error

		// Initialize the appropriate provider
		switch provider {
		case "groq":
			// Get Groq API key
			if apiKey == "" {
				apiKey = GetConfig().GetGroqApiKey()
			}

			if apiKey == "" {
				envVar := GetConfig().General.ApiKeyGroqEnvVariable
				if envVar == "" {
					envVar = "GOCO_GROQ_KEY"
				}

				promptedKey, err := promptForApiKey(envVar, "Groq")
				if err != nil {
					return fmt.Errorf("failed to get Groq API key: %w", err)
				}
				apiKey = promptedKey
			}

			// Set default model for Groq if not specified
			if model == "" {
				model = "llama-3.3-70b-versatile"
			}

			aiProvider, err = providers.NewGroqProvider(ctx, apiKey, model)
			if err != nil {
				return &ProviderError{
					Provider: "groq",
					Message:  "failed to initialize provider",
					Err:      err,
				}
			}

		case "gemini":
			// Get Gemini API key
			if apiKey == "" {
				apiKey = GetConfig().GetGeminiApiKey()
			}

			if apiKey == "" {
				envVar := GetConfig().General.ApiKeyGeminiEnvVariable
				if envVar == "" {
					envVar = "GOCO_GEMINI_KEY"
				}

				promptedKey, err := promptForApiKey(envVar, "Gemini")
				if err != nil {
					return fmt.Errorf("failed to get Gemini API key: %w", err)
				}
				apiKey = promptedKey
			}

			// Set default model for Gemini if not specified
			if model == "" {
				model = "gemini-2.5-flash"
			}

			aiProvider, err = providers.NewGeminiProvider(ctx, apiKey, model)
			if err != nil {
				return &ProviderError{
					Provider: "gemini",
					Message:  "failed to initialize provider",
					Err:      err,
				}
			}

		default:
			return &ValidationError{
				Field:   "provider",
				Message: fmt.Sprintf("unsupported provider %q", provider),
				Help:    "supported providers: gemini, groq",
			}
		}

		if err := aiProvider.ValidateModel(ctx, model); err != nil {
			return &ValidationError{
				Field:   "model",
				Message: fmt.Sprintf("validation failed for model %q", model),
				Help:    "run 'goco models --provider <provider>' to list available models",
			}
		}

		gitStatus := exec.Command("git", "status")
		gitStatusOutput, err := gitStatus.Output()
		if err != nil {
			return &GitError{
				Command: "git status",
				Message: "failed to get repository status",
				Err:     err,
			}
		}

		if len(strings.TrimSpace(string(gitStatusOutput))) == 0 {
			return &GitError{
				Command: "git status",
				Message: "no changes detected",
				Err:     ErrGitRepository,
			}
		}

		// Get git diff
		var gitDiff *exec.Cmd
		// Use the single `staged` flag.
		useStaged := staged
		if useStaged {
			gitDiff = exec.Command("git", "diff", "--no-color", "--staged")
		} else {
			gitDiff = exec.Command("git", "diff", "--no-color")
		}

		gitDiffOutput, err := gitDiff.Output()
		if err != nil {
			return &GitError{
				Command: "git diff",
				Message: "failed to get diff",
				Err:     err,
			}
		}

		if verbose {
			// Show git status in a green box
			statusBox := statusBoxStyle.Render(string(gitStatusOutput))
			fmt.Println(statusHeaderStyle.Render("📊 Git Status"))
			fmt.Println(statusBox)

			// Show git diff in a blue box
			diffBox := diffBoxStyle.Render(string(gitDiffOutput))
			fmt.Println(diffHeaderStyle.Render("📝 Git Diff"))
			fmt.Println(diffBox)
		}

		// Start spinner during API call
		spinnerProgram := tea.NewProgram(newSpinnerModel("Generating commit message..."))

		// Run spinner in goroutine
		done := make(chan bool)
		go func() {
			spinnerProgram.Run()
			done <- true
		}()

		// Make API call
		commitMessage, err := aiProvider.GenerateCommitMessage(
			ctx,
			string(gitStatusOutput),
			string(gitDiffOutput),
			customInstructions,
		)

		// Stop spinner
		spinnerProgram.Send("done")
		spinnerProgram.Quit()
		<-done // Wait for spinner to finish

		if err != nil {
			return &APIError{
				Message: "failed to generate commit message",
				Err:     err,
			}
		}

		fmt.Println(commitMessageHeaderStyle.Render("✅ Generated Commit Message"))
		fmt.Println(commitMessageBoxStyle.Render(commitMessage))

		if edit {
			fmt.Println(titleStyle.Render("✏️  Editing Commit Message"))
			editedMessage, err := editCommitMessage(commitMessage)
			if err != nil {
				return fmt.Errorf("failed to edit commit message: %w", err)
			}
			commitMessage = editedMessage
			fmt.Println(commitMessageHeaderStyle.Render("✅ Final Commit Message"))
			fmt.Println(commitMessageBoxStyle.Render(commitMessage))
		}

		// Only update the index for modified tracked files when not explicitly
		// requesting to use the already staged changes. If the user passed the
		// --staged flag we must NOT modify the index, otherwise we
		// risk including unstaged changes in the commit (this was the bug).
		useStaged = staged

		if !useStaged {
			if err := exec.Command("git", "add", "-u").Run(); err != nil {
				return &GitError{
					Command: "git add -u",
					Message: "failed to stage changes",
					Err:     err,
				}
			}
		} else if verbose {
			fmt.Println("Using staged changes only; skipping 'git add -u'")
		}

		// When using staged changes, we must ensure the commit only contains
		// files currently in the index. Otherwise `git commit` will include
		// unstaged modifications if `git add` was run earlier. To be explicit,
		// when staged is set we pass `--only` with a list of staged files.

		var final *exec.Cmd
		if useStaged {
			// Get the list of staged files
			stagedListCmd := exec.Command("git", "diff", "--name-only", "--cached")
			stagedOut, err := stagedListCmd.Output()
			if err != nil {
				return &GitError{
					Command: "git diff --name-only --cached",
					Message: "failed to list staged files",
					Err:     err,
				}
			}

			stagedFiles := strings.Fields(strings.TrimSpace(string(stagedOut)))
			if len(stagedFiles) == 0 {
				return &GitError{
					Command: "git diff --name-only --cached",
					Message: "no staged changes to commit",
					Err:     ErrGitRepository,
				}
			}

			// `git commit --only <path>...` commits only the specified paths
			// from the index. Build args as: commit -m <msg> --only -- <paths...>
			args := []string{"commit", "-m", commitMessage, "--only", "--"}
			args = append(args, stagedFiles...)
			final = exec.Command("git", args...)
		} else {
			final = exec.Command("git", "commit", "-m", commitMessage)
		}
		final.Stdout = os.Stdout
		final.Stderr = os.Stderr

		if err := final.Run(); err != nil {
			return &GitError{
				Command: "git commit",
				Message: "failed to commit changes",
				Err:     err,
			}
		}

		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&provider, "provider", "p", "", "AI provider to use (gemini or groq)")
	generateCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key for the selected provider")
	generateCmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (defaults: gemini-2.5-flash for Gemini, llama-3.3-70b-versatile for Groq)")
	generateCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	generateCmd.Flags().BoolVarP(&breakingChange, "breaking-change", "b", false, "Mark commit as breaking change")
	// These flags are currently unused by the CLI logic and are deprecated.
	// Keep them present for backwards compatibility but mark as deprecated so
	// users see a deprecation notice in help output.
	if f := generateCmd.Flags().Lookup("type"); f != nil {
		_ = f.Deprecated // access to avoid lint unused warning
		_ = generateCmd.Flags().MarkDeprecated("type", "flag is unused and will be removed in a future release")
	}
	if f := generateCmd.Flags().Lookup("breaking-change"); f != nil {
		_ = f.Deprecated
		_ = generateCmd.Flags().MarkDeprecated("breaking-change", "flag is unused and will be removed in a future release")
	}
	// Register only the correctly spelled --staged flag (shorthand -s).
	generateCmd.Flags().BoolVarP(&staged, "staged", "s", false, "staged changes")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output including prompts")
	generateCmd.Flags().StringVarP(&customInstructions, "custom-instructions", "c", "", "Custom instructions to add to the prompt")
	generateCmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit the commit message before committing")

	rootCmd.AddCommand(generateCmd)
}
