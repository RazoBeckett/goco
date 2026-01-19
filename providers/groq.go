package providers

import (
	"context"
	"fmt"
	"slices"

	"github.com/conneroisu/groq-go"
)

// GroqProvider implements the Provider interface for Groq
type GroqProvider struct {
	client *groq.Client
	model  string
}

// NewGroqProvider creates a new Groq provider
func NewGroqProvider(ctx context.Context, apiKey, model string) (*GroqProvider, error) {
	client, err := groq.NewClient(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create groq client: %w", err)
	}

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

	resp, err := g.client.ChatCompletion(ctx, groq.ChatCompletionRequest{
		Model: groq.ChatModel(g.model),
		Messages: []groq.ChatCompletionMessage{
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

// ListModels lists available Groq models
func (g *GroqProvider) ListModels(ctx context.Context) ([]string, error) {
	// Return a list of text generation models suitable for commit message generation
	// Updated to exclude deprecated models and non-text models (TTS, STT)
	// Based on https://console.groq.com/docs/deprecations
	return []string{
		// Compound/Systems (Agentic AI) - Production
		"groq/compound",
		"groq/compound-mini",
		// Production Text Models
		"llama-3.1-8b-instant",
		"llama-3.3-70b-versatile",
		"mixtral-8x7b-32768",
		"openai/gpt-oss-120b",
		"openai/gpt-oss-20b",
		// Preview Text Models (may be deprecated with short notice)
		"meta-llama/llama-4-maverick-17b-128e-instruct",
		"meta-llama/llama-4-scout-17b-16e-instruct",
		"moonshotai/kimi-k2-instruct-0905",
		"qwen/qwen3-32b",
	}, nil
}

// ValidateModel validates that a model is available
func (g *GroqProvider) ValidateModel(ctx context.Context, model string) error {
	models, _ := g.ListModels(ctx)

	if !slices.Contains(models, model) {
		return fmt.Errorf("model %s not available", model)
	}

	return nil
}
