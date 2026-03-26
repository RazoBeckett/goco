package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/razobeckett/goco/internal/ai"
	"github.com/spf13/cobra"
)

type modelsOptions struct {
	provider string
	apiKey   string
}

func newModelsCmd(deps dependencies) *cobra.Command {
	opts := &modelsOptions{}

	cmd := &cobra.Command{
		Use:     "models",
		Short:   "List available AI models",
		Long:    "List all available AI models for the selected provider.",
		GroupID: "inspect",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runModels(cmd, deps, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "", "AI provider to list models for (gemini or groq)")
	cmd.Flags().StringVarP(&opts.apiKey, "api-key", "k", "", "API key for the selected provider")
	return cmd
}

func runModels(cmd *cobra.Command, deps dependencies, opts *modelsOptions) error {
	ctx := cmd.Context()

	cfg, err := deps.configLoader.Load()
	if err != nil {
		return fmt.Errorf("load config %q: %w", deps.configLoader.Path(), err)
	}

	providerName := opts.provider
	if providerName == "" {
		providerName = cfg.DefaultProviderName()
	}
	if providerName != ai.ProviderGemini && providerName != ai.ProviderGroq {
		return fmt.Errorf("invalid provider %q; supported providers: gemini, groq", providerName)
	}

	apiKey := opts.apiKey
	if apiKey == "" {
		apiKey = cfg.APIKey(providerName)
	}
	if apiKey == "" {
		apiKey, err = promptForAPIKey(cfg.APIKeyEnv(providerName), providerDisplayName(providerName))
		if err != nil {
			return err
		}
	}

	provider, err := ai.NewProvider(ctx, providerName, apiKey, "")
	if err != nil {
		return err
	}

	models, err := fetchModelsWithSpinner(ctx, provider)
	if err != nil {
		return err
	}

	fmt.Println(modelProviderStyle.Render(fmt.Sprintf("Available %s Models (%d found)", providerDisplayName(providerName), len(models))))
	fmt.Println()
	for _, model := range models {
		fmt.Println(modelItemStyle.Render("• " + model))
	}
	fmt.Println()
	fmt.Println(noteStyle.Render(fmt.Sprintf("Use --model with %s generate to pick a specific model.", cmd.Root().Name())))

	return nil
}

func fetchModelsWithSpinner(ctx context.Context, provider ai.Provider) ([]string, error) {
	program := tea.NewProgram(newSpinnerModel(fmt.Sprintf("Fetching %s models...", providerDisplayName(provider.Name()))))
	resultCh := make(chan struct {
		models []string
		err    error
	}, 1)

	go func() {
		models, err := provider.ListModels(ctx)
		resultCh <- struct {
			models []string
			err    error
		}{models: models, err: err}

		if err != nil {
			program.Send(spinnerErrorMsg{err: err})
			return
		}
		program.Send(spinnerStringListMsg{items: models})
	}()

	if _, err := program.Run(); err != nil {
		return nil, fmt.Errorf("run spinner: %w", err)
	}

	result := <-resultCh
	if result.err != nil {
		return nil, fmt.Errorf("list models: %w", result.err)
	}

	return result.models, nil
}
