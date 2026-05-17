package agent

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentConfig defines an agent in YAML.
type AgentConfig struct {
	Name            string            `yaml:"name"`
	Role            string            `yaml:"role"`
	Goal            string            `yaml:"goal"`
	Backstory       string            `yaml:"backstory"`
	SystemPrompt    string            `yaml:"system_prompt"`
	MaxIterations   int               `yaml:"max_iterations"`
	Tools           []string          `yaml:"tools"`
	Skills          []string          `yaml:"skills"`
	LLMConfig       *AgentLLMConfig   `yaml:"llm_config"`
	ResponseFormat  *ResponseFormatConfig `yaml:"response_format"`
}

// AgentLLMConfig holds per-agent LLM settings.
type AgentLLMConfig struct {
	Temperature     float64 `yaml:"temperature"`
	TopP            float64 `yaml:"top_p"`
	MaxTokens       int     `yaml:"max_tokens"`
	EnableReasoning bool    `yaml:"enable_reasoning"`
	ReasoningBudget int     `yaml:"reasoning_budget"`
}

// ResponseFormatConfig holds structured output config.
type ResponseFormatConfig struct {
	Type         string                 `yaml:"type"`
	SchemaName   string                 `yaml:"schema_name"`
	SchemaDef    map[string]interface{} `yaml:"schema_definition"`
}

// TaskConfig defines a task in YAML.
type TaskConfig struct {
	Description    string                `yaml:"description"`
	ExpectedOutput string                `yaml:"expected_output"`
	Agent          string                `yaml:"agent"`
	OutputFile     string                `yaml:"output_file"`
	ResponseFormat *ResponseFormatConfig `yaml:"response_format"`
}

// LoadAgentConfigsFromFile loads agent configurations from a YAML file.
func LoadAgentConfigsFromFile(path string) (map[string]AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read agent config file: %w", err)
	}

	// Expand environment variables
	content := os.ExpandEnv(string(data))

	var configs map[string]AgentConfig
	if err := yaml.Unmarshal([]byte(content), &configs); err != nil {
		return nil, fmt.Errorf("unmarshal agent config: %w", err)
	}

	return configs, nil
}

// LoadTaskConfigsFromFile loads task configurations from a YAML file.
func LoadTaskConfigsFromFile(path string) (map[string]TaskConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task config file: %w", err)
	}

	content := os.ExpandEnv(string(data))

	var configs map[string]TaskConfig
	if err := yaml.Unmarshal([]byte(content), &configs); err != nil {
		return nil, fmt.Errorf("unmarshal task config: %w", err)
	}

	return configs, nil
}

// SaveAgentConfigsToFile writes agent configs to a YAML file.
func SaveAgentConfigsToFile(configs map[string]AgentConfig, path string) error {
	data, err := yaml.Marshal(configs)
	if err != nil {
		return fmt.Errorf("marshal agent configs: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// SaveTaskConfigsToFile writes task configs to a YAML file.
func SaveTaskConfigsToFile(configs map[string]TaskConfig, path string) error {
	data, err := yaml.Marshal(configs)
	if err != nil {
		return fmt.Errorf("marshal task configs: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// BuildSystemPromptFromConfig builds a system prompt from agent config.
func BuildSystemPromptFromConfig(cfg AgentConfig) string {
	var parts []string

	if cfg.Role != "" {
		parts = append(parts, fmt.Sprintf("Role: %s", cfg.Role))
	}
	if cfg.Goal != "" {
		parts = append(parts, fmt.Sprintf("Goal: %s", cfg.Goal))
	}
	if cfg.Backstory != "" {
		parts = append(parts, fmt.Sprintf("Backstory: %s", cfg.Backstory))
	}
	if cfg.SystemPrompt != "" {
		parts = append(parts, cfg.SystemPrompt)
	}

	return strings.Join(parts, "\n\n")
}
