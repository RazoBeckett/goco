package providers

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/genai"
)

func TestListModels_Gemini25FlashIncluded(t *testing.T) {
	original := geminiListModelsFunc
	defer func() { geminiListModelsFunc = original }()

	geminiListModelsFunc = func(g *GeminiProvider, ctx context.Context) ([]*genai.Model, error) {
		return []*genai.Model{
			{Name: "models/gemini-1.5-flash"},
			{Name: "models/gemini-2.5-flash"},
			{Name: "models/gemini-2.5-flash-raw"},
			{Name: "models/gemini-1.5-pro"},
			{Name: "models/gemini-flash"},
			{Name: "models/gemini-pro"},
			{Name: "models/some-other-model"},
		}, nil
	}

	provider := &GeminiProvider{}
	result, err := provider.ListModels(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"gemini-1.5-flash",
		"gemini-2.5-flash",
		"gemini-2.5-flash-raw",
		"gemini-1.5-pro",
		"gemini-flash",
		"gemini-pro",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestListModels_FallbackWhenEmpty(t *testing.T) {
	original := geminiListModelsFunc
	defer func() { geminiListModelsFunc = original }()

	geminiListModelsFunc = func(g *GeminiProvider, ctx context.Context) ([]*genai.Model, error) {
		return []*genai.Model{
			{Name: "models/custom-model-1"},
			{Name: "models/custom-model-2"},
		}, nil
	}

	provider := &GeminiProvider{}
	result, err := provider.ListModels(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"custom-model-1",
		"custom-model-2",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v (fallback), got %v (no fallback)", expected, result)
	}
}
