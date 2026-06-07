package cli

import (
	"fmt"
	"os"

	"github.com/razobeckett/goco/internal/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type generateOptions struct {
	provider           string
	apiKey             string
	model              string
	customInstructions string
	newBranch          string
	staged             bool
	verbose            bool
	edit               bool
	noConfirm          bool
}

func newGenerateOptions() *generateOptions {
	return &generateOptions{}
}

func newGenerateCmd(deps dependencies) *cobra.Command {
	opts := newGenerateOptions()

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate and optionally apply a Conventional Commit",
		GroupID: "main",
		Args:    cobra.NoArgs,
		Example: "  goco generate\n  goco generate --provider gemini --model gemini-2.5-flash\n  goco generate --staged --edit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runGenerate(cmd, deps, opts)
		},
	}

	bindGenerateFlags(cmd.Flags(), opts)
	return cmd
}

func bindGenerateFlags(fs *pflag.FlagSet, opts *generateOptions) {
	fs.StringVarP(&opts.provider, "provider", "p", "", "AI provider to use (gemini or groq)")
	fs.StringVarP(&opts.apiKey, "api-key", "k", "", "API key for the selected provider")
	fs.StringVarP(&opts.model, "model", "m", "", "Model to use (defaults to the provider's recommended model)")
	fs.BoolVarP(&opts.staged, "staged", "s", false, "Use staged changes instead of the working tree diff")
	fs.BoolVarP(&opts.verbose, "verbose", "V", false, "Show git status and diff before generating the commit")
	fs.BoolVarP(&opts.noConfirm, "yes", "y", false, "Skip confirmation and commit immediately")
	fs.StringVarP(&opts.customInstructions, "custom-instructions", "c", "", "Additional instructions to add to the AI prompt")
	fs.BoolVarP(&opts.edit, "edit", "e", false, "Open the generated commit message in your editor before committing")
	fs.StringVarP(&opts.newBranch, "branch", "B", "", "Create a new branch from the current branch before committing")
}

func runGenerate(cmd *cobra.Command, deps dependencies, opts *generateOptions) error {
	pipeline := NewPipeline(deps, opts)
	return pipeline.Run(cmd.Context())
}

func promptForAPIKey(envVar, providerName string) (string, error) {
	fmt.Println(titleStyle.Render(fmt.Sprintf("%s API Key Required", providerName)))
	apiKey, err := runAPIKeyPrompt(providerName, envVar)
	if err != nil {
		return "", fmt.Errorf("read API key: %w", err)
	}

	// Best-effort: make the key available to child processes this session.
	if setErr := os.Setenv(envVar, apiKey); setErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not set %s: %v\n", envVar, setErr)
	}

	fmt.Println(noteStyle.Render(fmt.Sprintf(
		"Set %s for this session. Add it to your shell profile to avoid the prompt next time.",
		envVar,
	)))
	fmt.Println()

	return apiKey, nil
}

func confirmCommit() (bool, error) {
	return runConfirmPrompt("Proceed with this commit?")
}

func providerDisplayName(provider string) string {
	switch provider {
	case ai.ProviderGroq:
		return "Groq"
	default:
		return "Gemini"
	}
}
