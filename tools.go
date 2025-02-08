package aichat

import (
	"encoding/json"
	"fmt"
)

// Tool represents a single tool
type Tool struct {
	Type     string   `yaml:"type" json:"type"`
	Function Function `yaml:"function" json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Arguments field is used by LLM to call tools
	Arguments string `yaml:"arguments,omitempty" json:"arguments,omitempty"`

	// Parameters field is provides available tool calling schema to LLM
	Parameters Parameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// Parameters defines the structure of function parameters
type Parameters struct {
	Type       string              `yaml:"type" json:"type"`
	Properties map[string]Property `yaml:"properties" json:"properties"`
	Required   []string            `yaml:"required" json:"required"`
}

// Property contains the individual parameter definitions
type Property struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
}

// ArgumentsMap parses the Arguments JSON string into a map[string]interface{}
func (f *Function) ArgumentsMap() (map[string]interface{}, error) {
	if f.Arguments == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(f.Arguments), &result); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %v", err)
	}
	return result, nil
}
