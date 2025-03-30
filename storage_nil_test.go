package aichat_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

// TestStorageLoadNilMessage tests handling of nil messages during loading
func TestStorageLoadNilMessage(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()

	// Create a chat with an empty ID to test loading
	session := &aichat.Chat{Options: aichat.Options{S3: s3}}

	// Add test data with a valid message and a message with null Message field to the mock S3
	// This JSON has a message with null Message field to test the nil message handling
	s3.SetData("test-nil-message-key", []byte(`{"id":"nil-message-id","messages":[{"role":"user","content":"Valid message"}, {"message":null}]}`))

	// Test loading
	err := session.Load(ctx, "test-nil-message-key")
	assert.NoError(t, err, "Failed to load session with nil message")
	assert.Equal(t, "nil-message-id", session.ID, "ID not loaded correctly")
	// Should only have one message since the nil one should be skipped
	assert.Len(t, session.Messages, 1, "Messages not loaded correctly")
	assert.Equal(t, "user", session.Messages[0].Role, "Message role not loaded correctly")
	assert.Equal(t, "Valid message", session.Messages[0].ContentString(), "Message content not loaded correctly")
}
