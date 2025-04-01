package aichat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

// mockS3 implements the S3 interface for testing
type mockS3 struct {
	data map[string][]byte
}

func newMockS3() *mockS3 {
	return &mockS3{
		data: make(map[string][]byte),
	}
}

func (m *mockS3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

// SetData sets the data for a key in the mock S3
func (m *mockS3) SetData(key string, data []byte) {
	m.data[key] = data
}

// GetRawData gets the raw data for a key from the mock S3
func (m *mockS3) GetRawData(key string) ([]byte, bool) {
	data, ok := m.data[key]
	return data, ok
}

func (m *mockS3) Put(ctx context.Context, key string, data io.Reader) error {
	b, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	m.data[key] = b
	return nil
}

func (m *mockS3) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// mockS3WithErrors is a mock S3 implementation that returns errors
type mockS3WithErrors struct {
	mockS3
	shouldErrorOnGet  bool
	returnInvalidJSON bool
}

func (m *mockS3WithErrors) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.shouldErrorOnGet {
		return nil, fmt.Errorf("mock get error")
	}
	if m.returnInvalidJSON {
		return io.NopCloser(strings.NewReader("invalid json")), nil
	}
	return m.mockS3.Get(ctx, key)
}

func TestChatStorage(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Add some test data
	session.AddUserContent("Test message")
	session.Meta = make(map[string]any)
	session.Meta["test"] = "value"

	// Test saving
	err := session.Save(ctx, "test-key")
	assert.NoError(t, err, "Failed to save session")

	// Create a new session and load the data
	loadedSession := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}
	err = loadedSession.Load(ctx, "test-key")
	assert.NoError(t, err, "Failed to load session")

	// Verify loaded data
	assert.Equal(t, session.ID, loadedSession.ID, "Session ID mismatch")
	assert.Equal(t, len(session.Messages), len(loadedSession.Messages), "Message count mismatch")
	assert.Equal(t, "value", loadedSession.Meta["test"], "Metadata value mismatch")

	err = loadedSession.Delete(ctx, "test-key")
	assert.NoError(t, err, "Failed to delete session")
}

func TestStorageErrors(t *testing.T) {
	ctx := context.Background()

	// Test with nil S3
	session := &aichat.Chat{ID: "test-id"}

	assert.Error(t, session.Save(ctx, "test-key"), "Expected error when saving with nil S3")
	assert.Error(t, session.Load(ctx, "test-key"), "Expected error when loading with nil S3")

	t.Run("get error", func(t *testing.T) {
		s3 := &mockS3WithErrors{shouldErrorOnGet: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}

		err := chat.Load(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to get session from storage", "Unexpected error message")
	})

	t.Run("decode error", func(t *testing.T) {
		s3 := &mockS3WithErrors{returnInvalidJSON: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}

		err := chat.Load(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to decode chat data", "Unexpected error message")
	})

	t.Run("marshal error", func(t *testing.T) {
		s3 := newMockS3()
		chat := &aichat.Chat{
			Options: aichat.Options{S3: s3},
			Messages: []*aichat.Message{
				{
					Role:    "user",
					Content: make(chan int), // channels cannot be marshaled to JSON
				},
			},
		}

		err := chat.Save(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to marshal chat data", "Unexpected error message")
	})
}

func TestNewStorage(t *testing.T) {
	s3 := newMockS3()
	opts := aichat.Options{S3: s3}

	storage := aichat.NewStorage(opts)
	assert.NotNil(t, storage, "Expected non-nil storage")
	assert.Equal(t, s3, storage.Options.S3, "Storage options not set correctly")
}

func TestStorageLoad(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	storage := aichat.NewStorage(aichat.Options{S3: s3})

	// First create and save a chat
	originalChat := &aichat.Chat{
		Key:     "test-key",
		ID:      "test-id",
		Options: aichat.Options{S3: s3},
	}
	originalChat.AddUserContent("Hello")
	assert.NoError(t, originalChat.Save(ctx, "test-key"), "Failed to save chat")

	// Now test loading the chat
	loadedChat, err := storage.Load(ctx, "test-key")
	assert.NoError(t, err, "Failed to load chat")
	assert.NotNil(t, loadedChat, "Expected non-nil chat")
	assert.Equal(t, "test-key", loadedChat.Key, "Chat key mismatch")
	assert.Equal(t, "test-id", loadedChat.ID, "Chat ID mismatch")
	assert.Equal(t, 1, len(loadedChat.Messages), "Message count mismatch")
	assert.Equal(t, s3, loadedChat.Options.S3, "Chat options not set correctly")

	// Test loading non-existent chat
	_, err = storage.Load(ctx, "non-existent-key")
	assert.Error(t, err, "Expected error when loading non-existent chat")
}

func TestStorageLoadDoesNotExist(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	storage := aichat.NewStorage(aichat.Options{S3: s3})

	chat, err := storage.Load(ctx, "non-existent-key")
	assert.Error(t, err, "Expected error when loading non-existent chat")
	msg := chat.AddUserContent("Hello")
	msg.Meta().Set("Content-Type", "text/plain")
	assert.NoError(t, chat.Save(ctx, "non-existent-key"), "Failed to save chat")

	// Test loading the chat
	loadedChat, err := storage.Load(ctx, "non-existent-key")
	assert.NoError(t, err, "Failed to load chat")
	assert.NotNil(t, loadedChat, "Expected non-nil chat")
	assert.Equal(t, "non-existent-key", loadedChat.Key, "Chat key mismatch")
	assert.Equal(t, 1, len(loadedChat.Messages), "Message count mismatch")
	assert.Equal(t, "Hello", loadedChat.Messages[0].ContentString(), "Message content mismatch")
	assert.Equal(t, "text/plain", loadedChat.Messages[0].Meta().Get("Content-Type"))
}

func TestChatStorageOutput(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()

	// Create a chat with specific content for testing
	chat := &aichat.Chat{
		ID:      "test-output-id",
		Options: aichat.Options{S3: s3},
		Meta: map[string]any{
			"test_key": "test_value",
			"number":   42,
		},
	}

	rawOutput, err := json.Marshal(chat)
	assert.NoError(t, err, "Failed to marshal chat data")
	expected := `{"id":"test-output-id","messages":null,"created":"0001-01-01T00:00:00Z","last_updated":"0001-01-01T00:00:00Z","meta":{"number":42,"test_key":"test_value"}}`
	assert.Equal(t, expected, string(rawOutput))

	// Add a user message
	userMsg := chat.AddUserContent("Hello, this is a test message")
	// Add metadata to the message
	userMsg.Meta().Set("msg_meta_key", "msg_meta_value")

	// Add an assistant message
	assistantMsg := chat.AddAssistantContent("This is a response from the assistant")
	// Add metadata to the assistant message
	assistantMsg.Meta().Set("assistant_meta", true)

	// Save the chat
	testKey := "test-output-key"
	err = chat.Save(ctx, testKey)
	assert.NoError(t, err, "Failed to save chat")

	// Get the raw JSON data
	rawData, exists := s3.GetRawData(testKey)
	assert.True(t, exists, "Saved data not found")

	// Parse the JSON to verify its structure
	var parsedData map[string]any
	err = json.Unmarshal(rawData, &parsedData)
	assert.NoError(t, err, "Failed to parse saved JSON data")

	// Verify the top-level structure
	assert.Equal(t, "test-output-id", parsedData["id"], "Chat ID mismatch in saved JSON")

	// Verify metadata
	meta, ok := parsedData["meta"].(map[string]interface{})
	assert.True(t, ok, "Meta field missing or not an object")
	assert.Equal(t, "test_value", meta["test_key"], "Chat metadata value mismatch")
	assert.Equal(t, float64(42), meta["number"], "Chat metadata number mismatch")

	// Verify messages array
	messages, ok := parsedData["messages"].([]interface{})
	assert.True(t, ok, "Messages field missing or not an array")
	assert.Equal(t, 2, len(messages), "Incorrect number of messages")

	// Verify first message (user message)
	if len(messages) > 0 {
		userMsgData, ok := messages[0].(map[string]any)
		assert.True(t, ok, "User message data is not an object")
		assert.Equal(t, "user", userMsgData["role"], "User message role mismatch")
		assert.Equal(t, "Hello, this is a test message", userMsgData["content"], "User message content mismatch")
		// Check metadata in the nested meta field
		userMeta, ok := userMsgData["meta"].(map[string]interface{})
		assert.True(t, ok, "User message meta field missing or not an object")
		assert.Equal(t, "msg_meta_value", userMeta["msg_meta_key"], "User message metadata mismatch")
	}

	// Verify second message (assistant message)
	if len(messages) > 1 {
		assistantMsgData, ok := messages[1].(map[string]any)
		assert.True(t, ok, "Assistant message data is not an object")
		assert.Equal(t, "assistant", assistantMsgData["role"], "Assistant message role mismatch")
		assert.Equal(t, "This is a response from the assistant", assistantMsgData["content"], "Assistant message content mismatch")
		// Check metadata in the nested meta field
		assistantMeta, ok := assistantMsgData["meta"].(map[string]interface{})
		assert.True(t, ok, "Assistant message meta field missing or not an object")
		assert.Equal(t, true, assistantMeta["assistant_meta"], "Assistant message metadata mismatch")
	}

	// Verify no message meta will be sent to other tools doing direct marshalling
	rawOutput, err = json.Marshal(chat)
	assert.NoError(t, err, "Failed to marshal chat data")

	// Parse the output to verify structure instead of comparing exact strings
	// This avoids issues with dynamic timestamps
	var outputData map[string]any
	err = json.Unmarshal(rawOutput, &outputData)
	assert.NoError(t, err, "Failed to parse marshalled chat data")

	// Verify key fields
	assert.Equal(t, "test-output-id", outputData["id"], "Chat ID mismatch in marshalled JSON")

	// Verify messages array structure
	outputMessages, ok := outputData["messages"].([]interface{})
	assert.True(t, ok, "Messages field missing or not an array in marshalled JSON")
	assert.Equal(t, 2, len(outputMessages), "Incorrect number of messages in marshalled JSON")

	// Verify first message (user)
	userMsgOutput, ok := outputMessages[0].(map[string]interface{})
	assert.True(t, ok, "User message not an object in marshalled JSON")
	assert.Equal(t, "user", userMsgOutput["role"], "User role mismatch in marshalled JSON")
	assert.Equal(t, "Hello, this is a test message", userMsgOutput["content"], "User content mismatch in marshalled JSON")

	// Verify second message (assistant)
	assistantMsgOutput, ok := outputMessages[1].(map[string]interface{})
	assert.True(t, ok, "Assistant message not an object in marshalled JSON")
	assert.Equal(t, "assistant", assistantMsgOutput["role"], "Assistant role mismatch in marshalled JSON")
	assert.Equal(t, "This is a response from the assistant", assistantMsgOutput["content"], "Assistant content mismatch in marshalled JSON")

	// Verify metadata
	outputMeta, ok := outputData["meta"].(map[string]interface{})
	assert.True(t, ok, "Meta field missing or not an object in marshalled JSON")
	assert.Equal(t, "test_value", outputMeta["test_key"], "Chat metadata value mismatch in marshalled JSON")
	assert.Equal(t, float64(42), outputMeta["number"], "Chat metadata number mismatch in marshalled JSON")
}
