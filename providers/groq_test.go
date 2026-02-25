package providers

import (
	"context"
	"os"
	"testing"

	"github.com/algolyzer/groq-go"
)

// TestValidateModel_Success tests that ValidateModel succeeds with a valid model.
// Note: This requires a valid API key and makes a real API call.
// Set GOCO_GROQ_KEY environment variable to run this test.
func TestValidateModel_Success(t *testing.T) {
	apiKey := os.Getenv("GOCO_GROQ_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GOCO_GROQ_KEY not set")
	}

	ctx := context.Background()
	provider, err := NewGroqProvider(ctx, apiKey, "llama-3.3-70b-versatile")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Act - use a known valid model
	err = provider.ValidateModel(ctx, "llama-3.3-70b-versatile")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for valid model, got: %v", err)
	}
}

// TestValidateModel_InvalidModel tests that ValidateModel fails for invalid models.
// Note: This requires a valid API key and makes a real API call.
func TestValidateModel_InvalidModel(t *testing.T) {
	apiKey := os.Getenv("GOCO_GROQ_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GOCO_GROQ_KEY not set")
	}

	ctx := context.Background()
	provider, err := NewGroqProvider(ctx, apiKey, "llama-3.3-70b-versatile")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Act - use an invalid model
	err = provider.ValidateModel(ctx, "invalid-model-name")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid model, got nil")
	}
}

// TestListModels_Success tests that ListModels returns models from the API.
// Note: This requires a valid API key and makes a real API call.
func TestListModels_Success(t *testing.T) {
	apiKey := os.Getenv("GOCO_GROQ_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GOCO_GROQ_KEY not set")
	}

	ctx := context.Background()
	provider, err := NewGroqProvider(ctx, apiKey, "llama-3.3-70b-versatile")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Act
	models, err := provider.ListModels(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Expected models to be returned, got none")
	}

	// Verify we got some expected models (optional sanity check)
	found := false
	for _, m := range models {
		if m != "" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Expected at least one valid model ID")
	}
}

// TestValidateModel_ListModelsError tests error handling when ListModels fails.
func TestValidateModel_ListModelsError(t *testing.T) {
	// Arrange: replace groqListModelsFunc to simulate an error
	original := groqListModelsFunc
	defer func() { groqListModelsFunc = original }()

	groqListModelsFunc = func(g *GroqProvider, ctx context.Context) (*groq.ModelListResponse, error) {
		return nil, context.Canceled
	}

	gp := &GroqProvider{}

	// Act
	err := gp.ValidateModel(context.Background(), "llama-3.3-70b-versatile")

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
