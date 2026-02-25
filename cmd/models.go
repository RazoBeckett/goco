package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/razobeckett/goco/providers"
	"github.com/spf13/cobra"
)

var (
	modelListProvider string

	modelHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true).
				Padding(0, 1)

	modelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA")).
			PaddingLeft(2)

	modelProviderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)
)

// modelsSpinnerModel is used to show a loading spinner while fetching models
type modelsSpinnerModel struct {
	spinner  spinner.Model
	message  string
	done     bool
	models   []string
	err      error
	provider string
}

func (m modelsSpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m modelsSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case modelsLoadedMsg:
		m.models = msg.models
		m.done = true
		return m, tea.Quit
	case errorMsg:
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m modelsSpinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

type modelsLoadedMsg struct {
	models []string
}

type errorMsg struct {
	err error
}

func newModelsSpinnerModel(message string) modelsSpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	return modelsSpinnerModel{
		spinner: s,
		message: message,
	}
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available AI models",
	Long:  `List all available AI models for the selected provider (gemini or groq).`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Use flag value for provider if provided, otherwise get from config
		if modelListProvider == "" {
			modelListProvider = GetConfig().GetDefaultProvider()
		}

		var aiProvider providers.Provider
		var err error
		var providerDisplayName string

		// Initialize the appropriate provider
		switch modelListProvider {
		case "groq":
			providerDisplayName = "Groq"
			// Get Groq API key
			apiKey := GetConfig().GetGroqApiKey()
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

			aiProvider, err = providers.NewGroqProvider(ctx, apiKey, "llama-3.3-70b-versatile")
			if err != nil {
				log.Fatalf("Failed to create Groq provider: %v", err)
			}

		case "gemini":
			providerDisplayName = "Gemini"
			// Get Gemini API key
			apiKey := GetConfig().GetGeminiApiKey()
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

			aiProvider, err = providers.NewGeminiProvider(ctx, apiKey, "gemini-2.5-flash")
			if err != nil {
				log.Fatalf("Failed to create Gemini provider: %v", err)
			}

		default:
			log.Fatalf("Unsupported provider: %s (supported: gemini, groq)", modelListProvider)
		}

		// Show spinner while fetching models
		spinnerProgram := tea.NewProgram(newModelsSpinnerModel(fmt.Sprintf("Fetching %s models...", providerDisplayName)))

		// Fetch models in goroutine
		var models []string
		var fetchErr error
		done := make(chan struct{})

		go func() {
			models, fetchErr = aiProvider.ListModels(ctx)
			if fetchErr != nil {
				spinnerProgram.Send(errorMsg{err: fetchErr})
			} else {
				spinnerProgram.Send(modelsLoadedMsg{models: models})
			}
			close(done)
		}()

		// Run spinner
		if _, err := spinnerProgram.Run(); err != nil {
			log.Fatalf("Failed to run spinner: %v", err)
		}

		// Wait for fetch to complete
		<-done

		if fetchErr != nil {
			log.Fatalf("Failed to list models: %v", fetchErr)
		}

		// Display models
		fmt.Println(modelProviderStyle.Render(fmt.Sprintf("📋 Available %s Models (%d found)", providerDisplayName, len(models))))
		fmt.Println()

		for _, m := range models {
			fmt.Println(modelItemStyle.Render(fmt.Sprintf("• %s", m)))
		}

		fmt.Println()
		fmt.Println(noteStyle.Render(fmt.Sprintf("Use --model flag to specify a model: goco generate --provider %s --model <model-name>", modelListProvider)))
	},
}

func init() {
	modelsCmd.Flags().StringVarP(&modelListProvider, "provider", "p", "", "AI provider to list models for (gemini or groq)")

	rootCmd.AddCommand(modelsCmd)
}
