package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
)

// Client generates commit messages with one AI provider.
type Client struct {
	name string
	lm   provider.LanguageModel
}

// New builds a Client for the named provider. The name must come from the
// registry. An empty model falls back to the provider's default.
func New(name, model, apiKey string) (*Client, error) {
	spec, ok := Lookup(name)
	if !ok {
		return nil, fmt.Errorf("unsupported provider %q; supported providers: %s", name, SupportedList())
	}

	if model == "" {
		model = spec.DefaultModel
	}

	return &Client{
		name: name,
		lm:   factories[name](model, apiKey),
	}, nil
}

// Name returns the provider name.
func (c *Client) Name() string { return c.name }

// GenerateCommitMessage asks the provider for a Conventional Commit message.
// The goai client retries transient API failures with exponential backoff.
func (c *Client) GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions, recentLog string) (string, error) {
	result, err := goai.GenerateText(
		ctx,
		c.lm,
		goai.WithSystem(systemPrompt),
		goai.WithPrompt(buildUserPrompt(gitStatus, gitDiff, customInstructions, recentLog)),
		goai.WithTemperature(0.2),
		goai.WithMaxRetries(2),
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", c.name, err)
	}

	return strings.TrimSpace(result.Text), nil
}
