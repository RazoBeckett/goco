package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	DefaultGeminiAPIKeyEnv = "GOCO_GEMINI_KEY"
	DefaultGroqAPIKeyEnv   = "GOCO_GROQ_KEY"
	DefaultProvider        = "gemini"
)

type General struct {
	GeminiAPIKeyEnv string `toml:"api_key_gemini_env_variable"`
	GroqAPIKeyEnv   string `toml:"api_key_groq_env_variable"`
	DefaultProvider string `toml:"default_provider"`
}

type Config struct {
	General General `toml:"General"`
}

type Loader struct {
	path string
}

func NewLoader() *Loader {
	return &Loader{path: configPath()}
}

func (l *Loader) Path() string {
	return l.path
}

func (l *Loader) Load() (*Config, error) {
	cfg := &Config{
		General: General{
			GeminiAPIKeyEnv: DefaultGeminiAPIKeyEnv,
			GroqAPIKeyEnv:   DefaultGroqAPIKeyEnv,
			DefaultProvider: DefaultProvider,
		},
	}

	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(l.path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) DefaultProviderName() string {
	if c.General.DefaultProvider == "" {
		return DefaultProvider
	}
	return c.General.DefaultProvider
}

func (c *Config) APIKeyEnv(provider string) string {
	switch provider {
	case "groq":
		if c.General.GroqAPIKeyEnv != "" {
			return c.General.GroqAPIKeyEnv
		}
		return DefaultGroqAPIKeyEnv
	default:
		if c.General.GeminiAPIKeyEnv != "" {
			return c.General.GeminiAPIKeyEnv
		}
		return DefaultGeminiAPIKeyEnv
	}
}

func (c *Config) APIKey(provider string) string {
	return os.Getenv(c.APIKeyEnv(provider))
}

func configPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "goco", "config.toml")
}
