package providers

import "context"

// Provider defines the interface for AI providers
type Provider interface {
	// GenerateCommitMessage generates a commit message based on git status and diff
	GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error)

	// ListModels returns available models for the provider
	ListModels(ctx context.Context) ([]string, error)

	// ValidateModel checks if a model is available
	ValidateModel(ctx context.Context, model string) error
}
