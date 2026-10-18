package ai

import "testing"

func TestSpecsAndFactoriesStayInSync(t *testing.T) {
	for name, spec := range specs {
		if _, ok := factories[name]; !ok {
			t.Errorf("provider %q has no factory", name)
		}
		if spec.DefaultModel == "" {
			t.Errorf("provider %q has no default model", name)
		}
		if spec.KeyEnv == "" {
			t.Errorf("provider %q has no key environment variable", name)
		}
	}

	for name := range factories {
		if _, ok := specs[name]; !ok {
			t.Errorf("factory %q has no spec", name)
		}
	}
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	if _, err := New("does-not-exist", "", "key"); err == nil {
		t.Fatal("expected an error for an unknown provider")
	}
}

func TestNewFillsDefaultModel(t *testing.T) {
	client, err := New("groq", "", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Name() != "groq" {
		t.Fatalf("expected provider %q, got %q", "groq", client.Name())
	}
}
