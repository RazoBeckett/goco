package ai

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/algolyzer/groq-go"
)

type GroqProvider struct {
	client *groq.Client
	model  string
}

func NewGroqProvider(_ context.Context, apiKey, model string) (*GroqProvider, error) {
	return &GroqProvider{
		client: groq.NewClient(apiKey),
		model:  model,
	}, nil
}

func (g *GroqProvider) Name() string {
	return ProviderGroq
}

func (g *GroqProvider) DefaultModel() string {
	return DefaultGroqModel
}

func (g *GroqProvider) GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error) {
	resp, err := g.client.CreateChatCompletion(ctx, groq.ChatCompletionRequest{
		Model: g.model,
		Messages: []groq.ChatMessage{
			{
				Role:    groq.RoleUser,
				Content: buildPrompt(gitStatus, gitDiff, customInstructions),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("Groq API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("Groq API returned no choices")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (g *GroqProvider) ListModels(ctx context.Context) ([]string, error) {
	resp, err := groqListModelsFunc(g, ctx)
	if err != nil {
		return nil, fmt.Errorf("list Groq models: %w", err)
	}

	models := make([]string, 0, len(resp.Data))
	for _, model := range resp.Data {
		if model.ID != "" {
			models = append(models, model.ID)
		}
	}

	return models, nil
}

var groqListModelsFunc = func(g *GroqProvider, ctx context.Context) (*groq.ModelListResponse, error) {
	return g.client.ListModels(ctx)
}

func (g *GroqProvider) ValidateModel(ctx context.Context, model string) error {
	models, err := g.ListModels(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(models, model) {
		return fmt.Errorf("model %q is not available for Groq", model)
	}

	return nil
}
