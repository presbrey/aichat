package aichat_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

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

func TestSessionStorage(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	session := aichat.NewChat("test-id", aichat.Options{
		S3: s3,
	})

	// Add some test data
	session.AddUserMessage("Test message")
	session.Metadata["test"] = "value"

	// Test saving
	err := session.Save(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Create a new session and load the data
	loadedSession := aichat.NewChat("test-id", aichat.Options{
		S3: s3,
	})
	err = loadedSession.Load(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify loaded data
	if loadedSession.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, loadedSession.ID)
	}

	if len(loadedSession.Messages) != len(session.Messages) {
		t.Errorf("Expected %d messages, got %d", len(session.Messages), len(loadedSession.Messages))
	}

	if loadedSession.Metadata["test"] != "value" {
		t.Errorf("Expected metadata value 'value', got %v", loadedSession.Metadata["test"])
	}

	err = loadedSession.Delete(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}
}

func TestSessionStorageErrors(t *testing.T) {
	ctx := context.Background()

	// Test with nil S3
	session := &aichat.Chat{ID: "test-id"}

	if err := session.Save(ctx, "test-key"); err == nil {
		t.Error("Expected error when saving with nil S3")
	}

	if err := session.Load(ctx, "test-key"); err == nil {
		t.Error("Expected error when loading with nil S3")
	}
}
