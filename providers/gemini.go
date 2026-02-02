package providers

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"google.golang.org/genai"
)

// Available Gemini models from models.dev (https://models.dev/api.json)
// As of Feb 2026, the following models are available for text generation:
//
// Latest models (2025 releases):
//   - gemini-3-flash-preview (2025-12-17) - Flash preview model
//   - gemini-3-pro-preview (2025-11-18) - Pro preview model
//   - gemini-2.5-flash (2025-03-20) - Flash model
//   - gemini-2.5-pro (2025-03-20) - Pro model
//   - gemini-flash-latest (2025-09-25) - Latest flash model
//   - gemini-flash-lite-latest (2025-09-25) - Latest flash lite model
//
// Flash series (lightweight, fast):
//   - gemini-2.0-flash-lite (2024-12-11) - Flash lite
//   - gemini-2.0-flash (2024-12-11) - Flash
//   - gemini-1.5-flash (2024-05-14) - Flash
//   - gemini-1.5-flash-8b (2024-10-03) - Flash 8B
//
// Pro series (more capable):
//   - gemini-1.5-pro (2024-02-15) - Pro
//
// Preview models:
//   - gemini-2.5-flash-preview-05-20 (2025-05-20)
//   - gemini-2.5-flash-preview-09-2025 (2025-09-25)
//   - gemini-2.5-flash-preview-04-17 (2025-04-17)
//   - gemini-2.5-pro-preview-05-06 (2025-05-06)
//   - gemini-2.5-pro-preview-06-05 (2025-06-05)
//   - gemini-2.5-flash-lite-preview-06-17 (2025-06-17)
//   - gemini-2.5-flash-lite-preview-09-2025 (2025-09-25)
//
// Specialized models:
//   - gemini-2.5-flash-image (2025-08-26) - Flash with image capabilities
//   - gemini-2.5-flash-image-preview (2025-08-26) - Flash image preview
//   - gemini-live-2.5-flash-preview-native-audio (2025-06-17) - Live with native audio
//   - gemini-live-2.5-flash (2025-09-01) - Live
//
// All models support text generation with context windows up to 1M tokens.
// Use gemini-2.5-flash or gemini-2.5-pro for best balance of speed and quality.
//
// Note: This list is for reference. The actual model list is fetched dynamically
// from the Google Gemini API at runtime via ListModels().

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
	models, err := geminiListModelsFunc(g, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var filtered []string
	re := regexp.MustCompile(`^gemini-`)
	for _, m := range models {
		name := strings.TrimPrefix(m.Name, "models/")
		if re.MatchString(name) {
			filtered = append(filtered, name)
		}
	}

	if len(filtered) == 0 {
		for _, m := range models {
			name := strings.TrimPrefix(m.Name, "models/")
			filtered = append(filtered, name)
		}
	}

	return filtered, nil
}

var geminiListModelsFunc = func(g *GeminiProvider, ctx context.Context) ([]*genai.Model, error) {
	models, err := g.client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	return models.Items, nil
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
