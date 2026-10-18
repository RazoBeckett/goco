package ai

import (
	"sort"
	"strings"

	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/anthropic"
	"github.com/zendev-sh/goai/provider/cerebras"
	"github.com/zendev-sh/goai/provider/deepseek"
	"github.com/zendev-sh/goai/provider/google"
	"github.com/zendev-sh/goai/provider/groq"
	"github.com/zendev-sh/goai/provider/mistral"
	"github.com/zendev-sh/goai/provider/openai"
	"github.com/zendev-sh/goai/provider/openrouter"
	"github.com/zendev-sh/goai/provider/xai"
)

// Spec describes one supported provider. Every provider here works with an API
// key alone; GoCo does not wire up cloud credentials or local endpoints.
type Spec struct {
	// DefaultModel is used when neither --model nor config sets one.
	DefaultModel string

	// KeyEnv is the conventional environment variable that holds the API key.
	KeyEnv string
}

// specs holds every provider GoCo can drive, keyed by the --provider value.
// Groq is the default.
var specs = map[string]Spec{
	"groq": {
		DefaultModel: "llama-3.3-70b-versatile",
		KeyEnv:       "GROQ_API_KEY",
	},
	"openai": {
		DefaultModel: "gpt-4o",
		KeyEnv:       "OPENAI_API_KEY",
	},
	"anthropic": {
		DefaultModel: "claude-sonnet-4-20250514",
		KeyEnv:       "ANTHROPIC_API_KEY",
	},
	"google": {
		DefaultModel: "gemini-2.5-flash",
		KeyEnv:       "GOOGLE_GENERATIVE_AI_API_KEY",
	},
	"deepseek": {
		DefaultModel: "deepseek-chat",
		KeyEnv:       "DEEPSEEK_API_KEY",
	},
	"openrouter": {
		DefaultModel: "openai/gpt-4o",
		KeyEnv:       "OPENROUTER_API_KEY",
	},
	"xai": {
		DefaultModel: "grok-3",
		KeyEnv:       "XAI_API_KEY",
	},
	"mistral": {
		DefaultModel: "mistral-large-latest",
		KeyEnv:       "MISTRAL_API_KEY",
	},
	"cerebras": {
		DefaultModel: "llama-3.3-70b",
		KeyEnv:       "CEREBRAS_API_KEY",
	},
}

// factories build the goai language model for each provider.
var factories = map[string]func(model, apiKey string) provider.LanguageModel{
	"groq":       groqModel,
	"openai":     openaiModel,
	"anthropic":  anthropicModel,
	"google":     googleModel,
	"deepseek":   deepseekModel,
	"openrouter": openrouterModel,
	"xai":        xaiModel,
	"mistral":    mistralModel,
	"cerebras":   cerebrasModel,
}

// Lookup returns the spec for a provider name.
func Lookup(name string) (Spec, bool) {
	spec, ok := specs[name]
	return spec, ok
}

// Names returns the supported provider names, sorted.
func Names() []string {
	names := make([]string, 0, len(specs))
	for name := range specs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SupportedList returns the provider names as a comma-separated list for errors.
func SupportedList() string {
	return strings.Join(Names(), ", ")
}

// Each factory passes the key only when present, so the provider can still
// fall back to its own environment variable when GoCo has no key.

func groqModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return groq.Chat(model)
	}
	return groq.Chat(model, groq.WithAPIKey(apiKey))
}

func openaiModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return openai.Chat(model)
	}
	return openai.Chat(model, openai.WithAPIKey(apiKey))
}

func anthropicModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return anthropic.Chat(model)
	}
	return anthropic.Chat(model, anthropic.WithAPIKey(apiKey))
}

func googleModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return google.Chat(model)
	}
	return google.Chat(model, google.WithAPIKey(apiKey))
}

func deepseekModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return deepseek.Chat(model)
	}
	return deepseek.Chat(model, deepseek.WithAPIKey(apiKey))
}

func openrouterModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return openrouter.Chat(model)
	}
	return openrouter.Chat(model, openrouter.WithAPIKey(apiKey))
}

func xaiModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return xai.Chat(model)
	}
	return xai.Chat(model, xai.WithAPIKey(apiKey))
}

func mistralModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return mistral.Chat(model)
	}
	return mistral.Chat(model, mistral.WithAPIKey(apiKey))
}

func cerebrasModel(model, apiKey string) provider.LanguageModel {
	if apiKey == "" {
		return cerebras.Chat(model)
	}
	return cerebras.Chat(model, cerebras.WithAPIKey(apiKey))
}
