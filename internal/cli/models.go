package cli

import (
	"context"
	"fmt"
	"sort"

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
		Long:    "List all available AI models for the selected provider. Uses the models.dev community registry by default — no API key required.",
		GroupID: "inspect",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runModels(cmd, deps, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "", "AI provider to list models for (gemini or groq)")
	cmd.Flags().StringVarP(&opts.apiKey, "api-key", "k", "", "API key for the selected provider (only needed if models.dev is unreachable)")
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

	displayName := providerDisplayName(providerName)

	// Stage 1: Try models.dev — fast, cached, no API key needed.
	models, source := tryModelsDev(ctx, providerName)
	if len(models) > 0 {
		displayModels(ctx, models, displayName, source, cmd.Root().Name())
		return nil
	}

	// Stage 2: models.dev unreachable — fall back to live API with spinner.
	apiKey := opts.apiKey
	if apiKey == "" {
		apiKey = cfg.APIKey(providerName)
	}
	if apiKey == "" {
		apiKey, err = promptForAPIKey(cfg.APIKeyEnv(providerName), displayName)
		if err != nil {
			return err
		}
	}

	provider, err := ai.NewProvider(ctx, providerName, apiKey, "")
	if err != nil {
		return err
	}

	models, err = fetchModelsWithSpinner(ctx, provider)
	if err != nil {
		return err
	}

	displayModels(ctx, models, displayName, "live API", cmd.Root().Name())
	return nil
}

// tryModelsDev attempts to get models from the models.dev registry cache.
// Returns (models, source_description). On failure, returns empty slice.
func tryModelsDev(ctx context.Context, providerName string) ([]string, string) {
	models, err := ai.ListModelsFromDev(providerName)
	if err != nil || len(models) == 0 {
		return nil, ""
	}

	// models.dev returns unsorted; stable-sort for consistent output.
	sort.Strings(models)

	return models, "models.dev registry"
}

// displayModels prints the model list with appropriate header and source note.
func displayModels(ctx context.Context, models []string, providerName, source, commandName string) {
	fmt.Println(modelProviderStyle.Render(
		fmt.Sprintf("Available %s Models (%d found)", providerName, len(models)),
	))
	fmt.Println()

	for _, model := range models {
		fmt.Println(modelItemStyle.Render("• " + model))
	}

	fmt.Println()
	fmt.Println(noteStyle.Render(
		fmt.Sprintf("Source: %s. Use --model with %s generate to pick a specific model.", source, commandName),
	))
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
