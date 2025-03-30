package aichat_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/presbrey/aichat"
)

func TestChatMarshalJSON(t *testing.T) {
	// Create a chat with some messages
	chat := &aichat.Chat{
		ID:          "test-marshal-id",
		Created:     time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
		LastUpdated: time.Date(2025, 3, 29, 12, 0, 0, 0, time.UTC),
	}
	chat.AddUserContent("Hello")
	chat.AddAssistantContent("Hi there")

	// Marshal the chat to JSON
	data, err := json.Marshal(chat)
	require.NoError(t, err)

	// Verify the JSON contains the expected fields
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "test-marshal-id", jsonMap["id"])
	assert.Equal(t, "2025-03-29T00:00:00Z", jsonMap["created"])
	// Don't check the exact last_updated time as it may be set automatically
	_, hasLastUpdated := jsonMap["last_updated"]
	assert.True(t, hasLastUpdated, "last_updated field should be present")
	
	// Check that messages are included
	messages, ok := jsonMap["messages"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, messages, 2)
}

func TestChatUnmarshalJSON(t *testing.T) {
	// Create a JSON string representing a chat
	jsonData := `{
		"id": "test-unmarshal-id",
		"created": "2025-03-29T00:00:00Z",
		"last_updated": "2025-03-29T12:00:00Z",
		"messages": [
			{
				"role": "user",
				"content": "Hello"
			},
			{
				"role": "assistant",
				"content": "Hi there"
			}
		],
		"meta": {
			"test_key": "test_value"
		}
	}`

	// Unmarshal the JSON into a chat
	var chat aichat.Chat
	err := json.Unmarshal([]byte(jsonData), &chat)
	require.NoError(t, err)

	// Verify the chat has the expected fields
	assert.Equal(t, "test-unmarshal-id", chat.ID)
	assert.Equal(t, time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC), chat.Created)
	assert.Equal(t, time.Date(2025, 3, 29, 12, 0, 0, 0, time.UTC), chat.LastUpdated)
	
	// Check that messages are included
	assert.Len(t, chat.Messages, 2)
	assert.Equal(t, "user", chat.Messages[0].Role)
	assert.Equal(t, "Hello", chat.Messages[0].ContentString())
	assert.Equal(t, "assistant", chat.Messages[1].Role)
	assert.Equal(t, "Hi there", chat.Messages[1].ContentString())
	
	// Check that metadata is included
	assert.Equal(t, "test_value", chat.Meta["test_key"])
}

func TestChatUnmarshalJSONError(t *testing.T) {
	// Create an invalid JSON string
	invalidJSON := `{"id": "test-id", "created": "invalid-date"}`

	// Try to unmarshal the invalid JSON
	var chat aichat.Chat
	err := json.Unmarshal([]byte(invalidJSON), &chat)
	assert.Error(t, err)
}
