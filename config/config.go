package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type General struct {
	ApiKeyGeminiEnvVariable string `toml:"api_key_gemini_env_variable"`
}

type Config struct {
	General General `toml:"General"`
}

func getConfigPath() string {
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

func LoadConfig() (*Config, error) {
	config := &Config{
		General: General{
			ApiKeyGeminiEnvVariable: "GOCO_GEMINI_KEY",
		},
	}

	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) GetGeminiApiKey() string {
	envVar := c.General.ApiKeyGeminiEnvVariable
	if envVar == "" {
		envVar = "GOCO_GEMINI_KEY"
	}
	return os.Getenv(envVar)
}

func (c *Config) CreateConfigFile() error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(c)
}
