package cmd

import (
	"context"
	"fmt"
	"log"
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
	apiKey             string
	provider           string
	model              string
	commitType         string
	breakingChange     bool
	stagged            bool
	verbose            bool
	customInstructions string
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

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a commit message using AI",

	Run: func(cmd *cobra.Command, args []string) {
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
					log.Fatalf("Failed to get API key: %v", err)
				}
				apiKey = promptedKey
			}

			// Set default model for Groq if not specified
			if model == "" {
				model = "llama-3.3-70b-versatile"
			}

			aiProvider, err = providers.NewGroqProvider(ctx, apiKey, model)
			if err != nil {
				log.Fatalf("Failed to create Groq provider: %v", err)
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
					log.Fatalf("Failed to get API key: %v", err)
				}
				apiKey = promptedKey
			}

			// Set default model for Gemini if not specified
			if model == "" {
				model = "gemini-2.5-flash"
			}

			aiProvider, err = providers.NewGeminiProvider(ctx, apiKey, model)
			if err != nil {
				log.Fatalf("Failed to create Gemini provider: %v", err)
			}

		default:
			log.Fatalf("Unsupported provider: %s (supported: gemini, groq)", provider)
		}

		// Validate the model
		if err := aiProvider.ValidateModel(ctx, model); err != nil {
			log.Fatalf("Model validation failed: %v", err)
		}

		// Get git status
		gitStatus := exec.Command("git", "status")
		gitStatusOutput, err := gitStatus.Output()
		if err != nil {
			log.Fatalf("Error getting git status: %v", err)
		}

		// Get git diff
		var gitDiff *exec.Cmd
		if stagged {
			gitDiff = exec.Command("git", "diff", "--no-color", "--staged")
		} else {
			gitDiff = exec.Command("git", "diff", "--no-color")
		}

		gitDiffOutput, err := gitDiff.Output()
		if err != nil {
			log.Fatalf("Error getting git diff: %v", err)
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
			log.Fatalf("AI API error: %v", err)
		}

		// Show the commit message in a beautiful green box
		fmt.Println(commitMessageHeaderStyle.Render("✅ Generated Commit Message"))
		fmt.Println(commitMessageBoxStyle.Render(commitMessage))

		if err := exec.Command("git", "add", "-u").Run(); err != nil {
			log.Fatalf("Failed to stage changes %v", err)
		}

		final := exec.Command("git", "commit", "-m", commitMessage)
		final.Stdout = os.Stdout
		final.Stderr = os.Stderr

		if err := final.Run(); err != nil {
			log.Fatalf("Failed to commit changes %v", err)
		}
	},
}

func init() {
	generateCmd.Flags().StringVarP(&provider, "provider", "p", "", "AI provider to use (gemini or groq)")
	generateCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key for the selected provider")
	generateCmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (defaults: gemini-2.5-flash for Gemini, llama-3.3-70b-versatile for Groq)")
	generateCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, chore, etc.)")
	generateCmd.Flags().BoolVarP(&breakingChange, "breaking-change", "b", false, "Mark commit as breaking change")
	generateCmd.Flags().BoolVarP(&stagged, "stagged", "s", false, "stagged changes")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output including prompts")
	generateCmd.Flags().StringVarP(&customInstructions, "custom-instructions", "c", "", "Custom instructions to add to the prompt")

	rootCmd.AddCommand(generateCmd)
}
