package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// DefaultProvider is used when config does not set one.
const DefaultProvider = "groq"

// Provider holds optional per-provider overrides. Everything here has a
// sensible default, so the file can stay empty.
type Provider struct {
	APIKeyEnvVariable string `toml:"api_key_env_variable"`
	Model             string `toml:"model"`
}

// General holds top-level settings.
type General struct {
	DefaultProvider string `toml:"default_provider"`
}

// Config is the parsed config file.
type Config struct {
	General   General             `toml:"General"`
	Providers map[string]Provider `toml:"providers"`
}

// Loader reads the config file from the XDG config directory.
type Loader struct {
	path string
}

// NewLoader returns a Loader for the default config path.
func NewLoader() *Loader {
	return &Loader{path: configPath()}
}

// Path returns the config file path.
func (l *Loader) Path() string {
	return l.path
}

// Load reads the config file. A missing file is not an error; defaults are used.
func (l *Loader) Load() (*Config, error) {
	cfg := &Config{
		General: General{DefaultProvider: DefaultProvider},
	}

	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(l.path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DefaultProviderName returns the configured provider, or the default.
func (c *Config) DefaultProviderName() string {
	if c.General.DefaultProvider == "" {
		return DefaultProvider
	}
	return c.General.DefaultProvider
}

// APIKeyEnv returns the environment variable that holds the key for a provider.
// A config override wins over the provider's conventional variable name.
func (c *Config) APIKeyEnv(providerName, fallback string) string {
	if p, ok := c.Providers[providerName]; ok && p.APIKeyEnvVariable != "" {
		return p.APIKeyEnvVariable
	}
	return fallback
}

// Model returns the configured model for a provider, or an empty string.
func (c *Config) Model(providerName string) string {
	if p, ok := c.Providers[providerName]; ok {
		return p.Model
	}
	return ""
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
