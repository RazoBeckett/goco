package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/razobeckett/goco/internal/ai"
	"github.com/razobeckett/goco/internal/git"
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
	commitType         string
	breakingChange     bool
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

	bindGenerateFlags(cmd.Flags(), opts, true)
	return cmd
}

func bindGenerateFlags(fs *pflag.FlagSet, opts *generateOptions, includeDeprecated bool) {
	fs.StringVarP(&opts.provider, "provider", "p", "", "AI provider to use (gemini or groq)")
	fs.StringVarP(&opts.apiKey, "api-key", "k", "", "API key for the selected provider")
	fs.StringVarP(&opts.model, "model", "m", "", "Model to use (defaults to the provider's recommended model)")
	fs.BoolVarP(&opts.staged, "staged", "s", false, "Use staged changes instead of the working tree diff")
	fs.BoolVarP(&opts.verbose, "verbose", "V", false, "Show git status and diff before generating the commit")
	fs.BoolVarP(&opts.noConfirm, "yes", "y", false, "Skip confirmation and commit immediately")
	fs.StringVarP(&opts.customInstructions, "custom-instructions", "c", "", "Additional instructions to add to the AI prompt")
	fs.BoolVarP(&opts.edit, "edit", "e", false, "Open the generated commit message in your editor before committing")
	fs.StringVarP(&opts.newBranch, "branch", "B", "", "Create a new branch from the current branch before committing")

	if includeDeprecated {
		fs.StringVarP(&opts.commitType, "type", "t", "", "Deprecated")
		_ = fs.MarkDeprecated("type", "flag is unused and will be removed")

		fs.BoolVarP(&opts.breakingChange, "breaking-change", "b", false, "Deprecated")
		_ = fs.MarkDeprecated("breaking-change", "flag is unused and will be removed")
	}
}

func runGenerate(cmd *cobra.Command, deps dependencies, opts *generateOptions) error {
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

	provider, err := ai.NewProvider(ctx, providerName, apiKey, opts.model)
	if err != nil {
		return err
	}

	modelName := opts.model
	if modelName == "" {
		modelName = provider.DefaultModel()
	}
	if err := provider.ValidateModel(ctx, modelName); err != nil {
		return fmt.Errorf("validate model %q: %w", modelName, err)
	}

	status, err := deps.repo.EnsureChanges(ctx)
	if err != nil {
		if err == git.ErrNoChanges {
			return fmt.Errorf("no changes detected; stage files or edit your working tree before running goco")
		}
		return err
	}

	diff, err := deps.repo.Diff(ctx, opts.staged)
	if err != nil {
		return fmt.Errorf("read git diff: %w", err)
	}

	if opts.verbose {
		fmt.Println(statusHeaderStyle.Render("Git Status"))
		fmt.Println(statusBoxStyle.Render(status))
		fmt.Println(diffHeaderStyle.Render("Git Diff"))
		fmt.Println(diffBoxStyle.Render(diff))
	}

	commitMessage, err := generateCommitMessage(ctx, provider, status, diff, opts.customInstructions)
	if err != nil {
		return err
	}

	fmt.Println(commitMessageHeaderStyle.Render("Generated Commit Message"))
	fmt.Println(commitMessageBoxStyle.Render(commitMessage))

	if opts.edit {
		fmt.Println(titleStyle.Render("Edit Commit Message"))

		commitMessage, err = editCommitMessage(commitMessage)
		if err != nil {
			return err
		}

		fmt.Println(commitMessageHeaderStyle.Render("Final Commit Message"))
		fmt.Println(commitMessageBoxStyle.Render(commitMessage))
	}

	if opts.newBranch != "" {
		currentBranch, err := deps.repo.CurrentBranch(ctx)
		if err != nil {
			return err
		}

		if err := deps.repo.CreateBranch(ctx, opts.newBranch); err != nil {
			return err
		}

		if opts.verbose {
			fmt.Printf("\nCreated and switched to %q from %q.\n\n", opts.newBranch, currentBranch)
		}
	}

	if !opts.noConfirm {
		confirmed, err := confirmCommit()
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println(noteStyle.Render("Commit cancelled."))
			return nil
		}
	}

	var stagedFiles []string
	if opts.staged {
		stagedFiles, err = deps.repo.StagedFiles(ctx)
		if err != nil {
			if err == git.ErrNoChanges {
				return fmt.Errorf("no staged changes to commit")
			}
			return err
		}
	} else {
		if err := deps.repo.StageTracked(ctx); err != nil {
			return err
		}
	}

	if err := deps.repo.Commit(ctx, commitMessage, stagedFiles); err != nil {
		return err
	}

	return nil
}

func generateCommitMessage(ctx context.Context, provider ai.Provider, status, diff, customInstructions string) (string, error) {
	program := tea.NewProgram(newSpinnerModel("Generating commit message..."))
	resultCh := make(chan struct {
		message string
		err     error
	}, 1)

	go func() {
		message, err := provider.GenerateCommitMessage(ctx, status, diff, customInstructions)
		resultCh <- struct {
			message string
			err     error
		}{message: message, err: err}

		if err != nil {
			program.Send(spinnerErrorMsg{err: err})
			return
		}
		program.Send(spinnerDoneMsg{})
	}()

	if _, err := program.Run(); err != nil {
		return "", fmt.Errorf("run spinner: %w", err)
	}

	result := <-resultCh
	if result.err != nil {
		return "", fmt.Errorf("generate commit message: %w", result.err)
	}
	if strings.TrimSpace(result.message) == "" {
		return "", fmt.Errorf("AI provider returned an empty commit message")
	}

	return strings.TrimSpace(result.message), nil
}

func promptForAPIKey(envVar, providerName string) (string, error) {
	fmt.Println(titleStyle.Render(fmt.Sprintf("%s API Key Required", providerName)))
	apiKey, err := runAPIKeyPrompt(providerName, envVar)
	if err != nil {
		return "", fmt.Errorf("read API key: %w", err)
	}

	_ = os.Setenv(envVar, apiKey)

	fmt.Println(noteStyle.Render(fmt.Sprintf("Set %s for this session. Add it to your shell profile to avoid the prompt next time.", envVar)))
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
