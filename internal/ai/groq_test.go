package ai

import (
	"context"
	"testing"

	"github.com/algolyzer/groq-go"
)

func TestGroqValidateModelListModelsError(t *testing.T) {
	original := groqListModelsFunc
	defer func() { groqListModelsFunc = original }()

	groqListModelsFunc = func(g *GroqProvider, ctx context.Context) (*groq.ModelListResponse, error) {
		return nil, context.Canceled
	}

	provider := &GroqProvider{}
	if err := provider.ValidateModel(context.Background(), DefaultGroqModel); err == nil {
		t.Fatal("expected error, got nil")
	}
}
