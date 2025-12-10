package cmd

import (
	"context"
	"fmt"
	"log"

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
			// For Groq, we don't need an API key to list models since they're hardcoded
			aiProvider, err = providers.NewGroqProvider(ctx, "dummy-key", "llama-3.3-70b-versatile")
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

		// Get available models
		models, err := aiProvider.ListModels(ctx)
		if err != nil {
			log.Fatalf("Failed to list models: %v", err)
		}

		// Display models
		fmt.Println(modelProviderStyle.Render(fmt.Sprintf("📋 Available %s Models", providerDisplayName)))
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
