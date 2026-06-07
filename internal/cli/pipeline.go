package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/razobeckett/goco/internal/ai"
	"github.com/razobeckett/goco/internal/git"
)

// ErrCancelled is a sentinel returned when the user declines the confirmation prompt.
// It signals a clean exit, not a failure.
var ErrCancelled = errors.New("commit cancelled")

var conventionalCommitRegex = regexp.MustCompile(
	`^(build|chore|ci|docs|feat|fix|perf|refactor|style|test)(\([^)]*\))?!?: .+`,
)

// Pipeline orchestrates the full generate flow as a sequence of cancellable stages.
// Each stage is independently testable and owns its lifecycle.
type Pipeline struct {
	deps dependencies
	opts *generateOptions

	// State accumulated across stages
	provider  ai.Provider
	modelName string
	status    string
	diff      string
	commitMsg string

	// Retry policy for transient AI failures
	maxRetries int
	retryDelay time.Duration
}

// NewPipeline creates a pipeline from the given dependencies and options.
func NewPipeline(deps dependencies, opts *generateOptions) *Pipeline {
	return &Pipeline{
		deps:       deps,
		opts:       opts,
		maxRetries: 2,
		retryDelay: 2 * time.Second,
	}
}

// Run advances through all pipeline stages in sequence.
// The outer context carries user cancellation (Ctrl+C); the pipeline
// wraps it with a hard timeout to prevent indefinite hangs.
func (p *Pipeline) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	stages := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"resolve", p.resolve},
		{"inspect", p.inspect},
		{"generate", p.generate},
		{"validate", p.validate},
		{"review", p.review},
		{"apply", p.apply},
	}

	for _, s := range stages {
		if err := s.fn(ctx); err != nil {
			if errors.Is(err, ErrCancelled) {
				return nil
			}
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	return nil
}

// --- Stage 1: Resolve config + provider + model ---

func (p *Pipeline) resolve(ctx context.Context) error {
	cfg, err := p.deps.configLoader.Load()
	if err != nil {
		return fmt.Errorf("load config %q: %w", p.deps.configLoader.Path(), err)
	}

	providerName := p.opts.provider
	if providerName == "" {
		providerName = cfg.DefaultProviderName()
	}
	if providerName != ai.ProviderGemini && providerName != ai.ProviderGroq {
		return fmt.Errorf("invalid provider %q; supported providers: gemini, groq", providerName)
	}

	apiKey := p.opts.apiKey
	if apiKey == "" {
		apiKey = cfg.APIKey(providerName)
	}
	if apiKey == "" {
		key, err := promptForAPIKey(cfg.APIKeyEnv(providerName), providerDisplayName(providerName))
		if err != nil {
			return err
		}
		apiKey = key
	}

	provider, err := ai.NewProvider(ctx, providerName, apiKey, p.opts.model)
	if err != nil {
		return err
	}

	modelName := p.opts.model
	if modelName == "" {
		modelName = provider.DefaultModel()
	} else if modelName != provider.DefaultModel() {
		// Only validate non-default models to save an API round-trip.
		if err := provider.ValidateModel(ctx, modelName); err != nil {
			return fmt.Errorf("validate model %q: %w", modelName, err)
		}
	}

	p.provider = provider
	p.modelName = modelName
	return nil
}

// --- Stage 2: Inspect git state ---

func (p *Pipeline) inspect(ctx context.Context) error {
	status, err := p.deps.repo.EnsureChanges(ctx)
	if err != nil {
		if err == git.ErrNoChanges {
			return fmt.Errorf("no changes detected; stage files or edit your working tree before running goco")
		}
		return err
	}

	diff, err := p.deps.repo.Diff(ctx, p.opts.staged)
	if err != nil {
		return fmt.Errorf("read git diff: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		if p.opts.staged {
			return fmt.Errorf("no staged changes to generate a commit from; stage files with `git add` first, or run without --staged to include working-tree changes")
		}
		return fmt.Errorf("no changes detected in the working tree; edit files before running goco")
	}

	p.status = status
	p.diff = diff

	if p.opts.verbose {
		fmt.Println(statusHeaderStyle.Render("Git Status"))
		fmt.Println(statusBoxStyle.Render(status))
		fmt.Println(diffHeaderStyle.Render("Git Diff"))
		fmt.Println(diffBoxStyle.Render(diff))
	}

	return nil
}

// --- Stage 3: Generate commit message via AI (with retry) ---

func (p *Pipeline) generate(ctx context.Context) error {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			delay := p.retryDelay * time.Duration(1<<(attempt-1))
			fmt.Fprintf(os.Stderr, "\nRetrying in %v (attempt %d/%d)...\n", delay, attempt+1, p.maxRetries+1)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		msg, err := p.spin(ctx, "Generating commit message...", func(ctx context.Context) (string, error) {
			return p.provider.GenerateCommitMessage(ctx, p.status, p.diff, p.opts.customInstructions)
		})
		if err == nil {
			if strings.TrimSpace(msg) == "" {
				return fmt.Errorf("AI provider returned an empty commit message")
			}
			p.commitMsg = strings.TrimSpace(msg)
			return nil
		}

		lastErr = err

		if !ai.IsTransient(err) {
			return fmt.Errorf("generate commit message: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf("generate commit message after %d retries: %w", p.maxRetries+1, lastErr)
}

// --- Stage 4: Validate the commit message ---

func (p *Pipeline) validate(_ context.Context) error {
	lines := strings.Split(p.commitMsg, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("commit message is empty")
	}

	subject := lines[0]
	if len(subject) > 72 {
		return fmt.Errorf(
			"commit subject is %d characters (max 72); use --edit to shorten it",
			len(subject),
		)
	}

	if !conventionalCommitRegex.MatchString(subject) {
		return fmt.Errorf(
			"commit subject %q does not match Conventional Commit format; expected <type>[scope]: <description>",
			subject,
		)
	}

	return nil
}

// --- Stage 5: Review — display, optional edit, confirm ---

func (p *Pipeline) review(ctx context.Context) error {
	fmt.Println(commitMessageHeaderStyle.Render("Generated Commit Message"))
	fmt.Println(commitMessageBoxStyle.Render(p.commitMsg))

	if p.opts.edit {
		fmt.Println(titleStyle.Render("Edit Commit Message"))

		edited, err := editCommitMessage(p.commitMsg)
		if err != nil {
			return err
		}
		p.commitMsg = edited

		fmt.Println(commitMessageHeaderStyle.Render("Final Commit Message"))
		fmt.Println(commitMessageBoxStyle.Render(p.commitMsg))

		// Re-validate after editing.
		if err := p.validate(ctx); err != nil {
			return fmt.Errorf("edited message is invalid: %w", err)
		}
	}

	if p.opts.noConfirm {
		return nil
	}

	confirmed, err := confirmCommit()
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println(noteStyle.Render("Commit cancelled."))
		return ErrCancelled
	}

	return nil
}

// --- Stage 6: Apply — branch, stage, commit ---

func (p *Pipeline) apply(ctx context.Context) error {
	if p.opts.newBranch != "" {
		currentBranch, err := p.deps.repo.CurrentBranch(ctx)
		if err != nil {
			return err
		}

		if err := p.deps.repo.CreateBranch(ctx, p.opts.newBranch); err != nil {
			return err
		}

		if p.opts.verbose {
			fmt.Printf("\nCreated and switched to %q from %q.\n\n", p.opts.newBranch, currentBranch)
		}
	}

	var stagedFiles []string
	var err error

	if p.opts.staged {
		stagedFiles, err = p.deps.repo.StagedFiles(ctx)
		if err != nil {
			if err == git.ErrNoChanges {
				return fmt.Errorf("no staged changes to commit")
			}
			return err
		}
	} else {
		if err := p.deps.repo.StageTracked(ctx); err != nil {
			return err
		}
	}

	if err := p.deps.repo.Commit(ctx, p.commitMsg, stagedFiles); err != nil {
		return err
	}

	return nil
}

// --- Spinner ---
// spin shows an animated spinner on stderr while fn executes.
// It respects ctx cancellation and cleans up on return.

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (p *Pipeline) spin(ctx context.Context, message string, fn func(context.Context) (string, error)) (string, error) {
	type result struct {
		msg string
		err error
	}

	done := make(chan result, 1)

	go func() {
		msg, err := fn(ctx)
		done <- result{msg, err}
	}()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case res := <-done:
			fmt.Fprint(os.Stderr, "\r\033[K") // clear spinner line
			return res.msg, res.err
		case <-ctx.Done():
			fmt.Fprint(os.Stderr, "\r\033[K")
			return "", ctx.Err()
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], message)
			i++
		}
	}
}
