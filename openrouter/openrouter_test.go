package openrouter

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/presbrey/aichat"
	"github.com/stretchr/testify/assert"
)

func TestOpenRouterToolCalls(t *testing.T) {
	b, err := os.ReadFile("openrouter-tool-calls.json")
	assert.NoError(t, err)

	resp := &Response{}
	assert.NoError(t, json.Unmarshal(b, resp))

	chat := &aichat.Chat{}
	chat.AddMessage(resp.Choices[0].Message)

	toolCallCount := 0
	chat.RangePendingToolCalls(func(tc *aichat.ToolCallContext) error {
		toolCallCount++
		args, err := tc.Arguments()
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"class":       "standard",
			"date":        "2025-05-25",
			"destination": "New York City",
			"origin":      "Boston",
			"passengers":  float64(1),
		}, args)
		return tc.Return(map[string]any{
			"foo": "bar",
		})
	})
	assert.Equal(t, 1, toolCallCount)

	assert.Equal(t, `{"foo":"bar"}`, chat.LastMessageByRole("tool").Content)
}
