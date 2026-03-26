package ai

import (
	"context"
	"fmt"
)

const (
	ProviderGemini = "gemini"
	ProviderGroq   = "groq"

	DefaultGeminiModel = "gemini-2.5-flash"
	DefaultGroqModel   = "llama-3.3-70b-versatile"
)

type Provider interface {
	Name() string
	DefaultModel() string
	GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error)
	ListModels(ctx context.Context) ([]string, error)
	ValidateModel(ctx context.Context, model string) error
}

func NewProvider(ctx context.Context, providerName, apiKey, model string) (Provider, error) {
	switch providerName {
	case ProviderGroq:
		return NewGroqProvider(ctx, apiKey, withDefault(model, DefaultGroqModel))
	case ProviderGemini:
		return NewGeminiProvider(ctx, apiKey, withDefault(model, DefaultGeminiModel))
	default:
		return nil, fmt.Errorf("unsupported provider %q (supported: gemini, groq)", providerName)
	}
}

func withDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
