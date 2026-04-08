package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AIProvider string
	AIModel    string
	APIKey     string
}

const (
	defaultProvider = "openai"
	defaultModel    = "gpt-5.1-codex"
)

func Load() Config {
	fileCfg, _ := LoadFromEnvFile(DefaultEnvFilePath())
	return Config{
		AIProvider: firstNonEmpty(os.Getenv("EDITPILOT_AI_PROVIDER"), fileCfg.AIProvider, defaultProvider),
		AIModel:    firstNonEmpty(os.Getenv("EDITPILOT_AI_MODEL"), fileCfg.AIModel, defaultModel),
		APIKey:     firstNonEmpty(os.Getenv("EDITPILOT_AI_API_KEY"), fileCfg.APIKey),
	}
}

func (c Config) ValidateForAI() error {
	if c.AIProvider == "" {
		return fmt.Errorf("AI provider is required")
	}
	if c.AIModel == "" {
		return fmt.Errorf("AI model is required")
	}
	return nil
}

func DefaultEnvFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".editpilot.env"
	}
	return filepath.Join(home, ".config", "editpilot", "config.env")
}

func InitEnvFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	content := "EDITPILOT_AI_PROVIDER=openai\nEDITPILOT_AI_MODEL=gpt-5.1-codex\nEDITPILOT_AI_API_KEY=\n"
	return os.WriteFile(path, []byte(content), 0o600)
}

func LoadFromEnvFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "EDITPILOT_AI_PROVIDER":
			cfg.AIProvider = value
		case "EDITPILOT_AI_MODEL":
			cfg.AIModel = value
		case "EDITPILOT_AI_API_KEY":
			cfg.APIKey = value
		}
	}
	return cfg, scanner.Err()
}

func SaveToEnvFile(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf("EDITPILOT_AI_PROVIDER=%s\nEDITPILOT_AI_MODEL=%s\nEDITPILOT_AI_API_KEY=%s\n", cfg.AIProvider, cfg.AIModel, cfg.APIKey)
	return os.WriteFile(path, []byte(content), 0o600)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
