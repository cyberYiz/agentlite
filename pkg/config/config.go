// Package config provides SDK-wide configuration.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the global SDK configuration.
type Config struct {
	LLM   LLMConfig
	Log   LogConfig
	Tools ToolsConfig
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	OpenAI OpenAIProviderConfig
}

// OpenAIProviderConfig holds OpenAI-specific settings.
type OpenAIProviderConfig struct {
	APIKey string
	Model  string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string // debug, info, warn, error
}

// ToolsConfig holds tool configuration.
type ToolsConfig struct {
	WebSearch WebSearchConfig
}

// WebSearchConfig holds web search tool settings.
type WebSearchConfig struct {
	GoogleAPIKey        string
	GoogleSearchEngineID string
	BraveAPIKey         string
}

// Get returns the configuration from environment variables.
func Get() *Config {
	return &Config{
		LLM: LLMConfig{
			OpenAI: OpenAIProviderConfig{
				APIKey: envOrDefault("OPENAI_API_KEY", ""),
				Model:  envOrDefault("OPENAI_MODEL", "gpt-4o-mini"),
			},
		},
		Log: LogConfig{
			Level: envOrDefault("LOG_LEVEL", "info"),
		},
		Tools: ToolsConfig{
			WebSearch: WebSearchConfig{
				GoogleAPIKey:        envOrDefault("GOOGLE_API_KEY", ""),
				GoogleSearchEngineID: envOrDefault("GOOGLE_SEARCH_ENGINE_ID", ""),
				BraveAPIKey:         envOrDefault("BRAVE_API_KEY", ""),
			},
		},
	}
}

// GetMaxIterations returns the configured max iterations or default.
func GetMaxIterations() int {
	v := os.Getenv("AGENT_MAX_ITERATIONS")
	if v == "" {
		return 15
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 15
	}
	return n
}

// GetRequirePlanApproval returns whether plan approval is required.
func GetRequirePlanApproval() bool {
	return strings.ToLower(os.Getenv("AGENT_REQUIRE_PLAN_APPROVAL")) == "true"
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
