package providers

import (
	"context"
	"fmt"
	"slices"

	"github.com/algolyzer/groq-go"
)

// GroqProvider implements the Provider interface for Groq
type GroqProvider struct {
	client *groq.Client
	model  string
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(ctx context.Context, apiKey, model string) (*GroqProvider, error) {
	client := groq.NewClient(apiKey)

	return &GroqProvider{
		client: client,
		model:  model,
	}, nil
}

// GenerateCommitMessage generates a commit message using Groq
func (g *GroqProvider) GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error) {
	referLink := "https://gist.githubusercontent.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13/raw/d7d529a329079616d47dcf100bd7d2d2c848e835/conventional-commits-cheatsheet.md"

	prompt := fmt.Sprintf(
		"Generate a Conventional Commit based strictly on the following:\n\n"+
			"Git Status:\n%s\n\n"+
			"Git Diff:\n%s\n\n"+
			"Before responding, you MUST:\n"+
			"- Read: %v\n"+
			"- ONLY output the commit message and description.\n"+
			"- DO NOT include markdown, code blocks, quotes, or any formatting.\n"+
			"- Output MUST be plain text only.\n"+
			"- Do not add extra explanations, notes, or commentary.\n"+
			"- The first line is the commit summary, the rest is the description.\n"+
			"- Follow Conventional Commit standards exactly.\n"+
			"- No extra lines before or after the commit message.\n",
		gitStatus,
		gitDiff,
		referLink,
	)

	if customInstructions != "" {
		prompt += fmt.Sprintf("\n\nAdditional Instructions:\n%s\n", customInstructions)
	}

	resp, err := g.client.CreateChatCompletion(ctx, groq.ChatCompletionRequest{
		Model: g.model,
		Messages: []groq.ChatMessage{
			{
				Role:    groq.RoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Groq API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq API")
	}

	return resp.Choices[0].Message.Content, nil
}

// ListModels lists available Groq models by fetching from the API.
// This allows users to see and select from all available models,
// not just a hardcoded subset.
func (g *GroqProvider) ListModels(ctx context.Context) ([]string, error) {
	resp, err := groqListModelsFunc(g, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list groq models: %w", err)
	}

	var models []string
	for _, model := range resp.Data {
		if model.ID != "" {
			models = append(models, model.ID)
		}
	}

	return models, nil
}

// groqListModelsFunc is a package-level indirection for ListModels to allow
// testing without making actual API calls.
var groqListModelsFunc = func(g *GroqProvider, ctx context.Context) (*groq.ModelListResponse, error) {
	return g.client.ListModels(ctx)
}

// ValidateModel validates that a model is available by checking against the API.
// This allows users to use any model that Groq supports, not just hardcoded ones.
func (g *GroqProvider) ValidateModel(ctx context.Context, model string) error {
	models, err := g.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list groq models: %w", err)
	}

	if !slices.Contains(models, model) {
		return fmt.Errorf("model %s not available", model)
	}

	return nil
}
