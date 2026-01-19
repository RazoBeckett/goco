package providers

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"google.golang.org/genai"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

// GenerateCommitMessage generates a commit message using Gemini
func (g *GeminiProvider) GenerateCommitMessage(ctx context.Context, gitStatus, gitDiff, customInstructions string) (string, error) {
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

	resp, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("Gemini API error: %w", err)
	}

	return resp.Text(), nil
}

// ListModels lists available Gemini models
func (g *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	models, err := g.client.Models.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var filtered []string
	re := regexp.MustCompile(`^gemini-\d+\.\d+`)
	for _, m := range models.Items {
		name := strings.TrimPrefix(m.Name, "models/")
		if re.MatchString(name) {
			filtered = append(filtered, name)
		}
	}

	return filtered, nil
}

// ValidateModel validates that a model is available
func (g *GeminiProvider) ValidateModel(ctx context.Context, model string) error {
	models, err := g.ListModels(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(models, model) {
		var b strings.Builder
		for _, m := range models {
			fmt.Fprintf(&b, "%s\n", m)
		}
		return fmt.Errorf("model not available\nAvailable Models:\n%s", b.String())
	}

	return nil
}
