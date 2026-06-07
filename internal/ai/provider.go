package ai

import (
	"context"
	"fmt"
	"net"
	"strings"
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

func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	transient := []string{
		"rate limit",
		"too many requests",
		"internal server error",
		"service unavailable",
		"timeout",
		"temporary",
		"connection reset",
		"broken pipe",
	}
	lower := strings.ToLower(msg)
	for _, keyword := range transient {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	return false
}
