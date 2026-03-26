package ai

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiProvider) Name() string {
	return ProviderGemini
}

func (g *GeminiProvider) DefaultModel() string {
	return DefaultGeminiModel
}

func (g *GeminiProvider) GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error) {
	resp, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(buildPrompt(gitStatus, gitDiff, customInstructions)),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("Gemini API error: %w", err)
	}

	return strings.TrimSpace(resp.Text()), nil
}

func (g *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	page, err := geminiListModelsFunc(g, ctx)
	if err != nil {
		return nil, fmt.Errorf("list Gemini models: %w", err)
	}

	var filtered []string
	re := regexp.MustCompile(`^gemini-`)

	for _, m := range page.Items {
		name := strings.TrimPrefix(m.Name, "models/")
		if re.MatchString(name) {
			filtered = append(filtered, name)
		}
	}

	if len(filtered) == 0 {
		for _, m := range page.Items {
			filtered = append(filtered, strings.TrimPrefix(m.Name, "models/"))
		}
	}

	return filtered, nil
}

var geminiListModelsFunc = func(g *GeminiProvider, ctx context.Context) (genai.Page[genai.Model], error) {
	return g.client.Models.List(ctx, nil)
}

func (g *GeminiProvider) ValidateModel(ctx context.Context, model string) error {
	models, err := g.ListModels(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(models, model) {
		return fmt.Errorf("model %q is not available for Gemini", model)
	}

	return nil
}
