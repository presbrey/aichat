package aichat

import (
	"encoding/json"
	"fmt"
)

// ToolCall represents a call to an external tool or function
type ToolCall struct {
	ID       string   `json:"id,omitempty"`
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

// RangePendingToolCalls iterates through messages to find and process tool calls that haven't received a response.
// It performs two passes: first to identify which tool calls have responses, then to process pending calls.
// The provided function is called for each pending tool call.
func (chat *Chat) RangePendingToolCalls(fn func(toolCall *ToolCallSession) error) error {
	// Create a map to track which tool calls have responses
	responded := make(map[string]bool)

	// First pass: mark which tool calls have responses
	for _, msg := range chat.Messages {
		if msg.ToolCallID != "" {
			responded[msg.ToolCallID] = true
		}
	}

	// Second pass: call the tool
	for _, msg := range chat.Messages {
		for _, call := range msg.ToolCalls {
			if !responded[call.ID] {
				if err := fn(&ToolCallSession{
					Chat:     chat,
					ToolCall: &call,
				}); err != nil {
					return err
				}
				responded[call.ID] = true
			}
		}
	}

	return nil
}

// ToolCallSession represents a tool call within a chat context, managing the lifecycle
// of a single tool invocation including its execution and response handling.
type ToolCallSession struct {
	ToolCall *ToolCall
	Chat     *Chat
}

// Name returns the name of the function
func (tcs *ToolCallSession) Name() string {
	return tcs.ToolCall.Function.Name
}

// Arguments returns the arguments to the function as a map
func (tcs *ToolCallSession) Arguments() (map[string]any, error) {
	return tcs.ToolCall.Function.ArgumentsMap()
}

// Return sends the result of the function call back to the chat
func (tcs *ToolCallSession) Return(result map[string]any) error {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}
	tcs.Chat.AddToolResponse(tcs.Name(), tcs.ToolCall.ID, string(jsonData))
	return nil
}
