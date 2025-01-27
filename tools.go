package aichat

import (
	"encoding/json"
	"fmt"
)

// ToolCall represents a call to an external tool or function
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// FunctionCall represents the function being called
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
