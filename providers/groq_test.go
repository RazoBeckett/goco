package providers

import (
	"context"
	"errors"
	"testing"
)

func TestValidateModel_ListModelsError(t *testing.T) {
	// Arrange: replace groqListModelsFunc to simulate an error
	original := groqListModelsFunc
	defer func() { groqListModelsFunc = original }()

	groqListModelsFunc = func(g *GroqProvider, ctx context.Context) ([]string, error) {
		return nil, errors.New("simulated list error")
	}

	gp := &GroqProvider{}

	// Act
	err := gp.ValidateModel(context.Background(), "llama-3.3-70b-versatile")

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, errors.New("simulated list error")) && err.Error() != "failed to list groq models: simulated list error" {
		t.Fatalf("unexpected error message: %v", err)
	}
}
