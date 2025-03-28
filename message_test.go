package aichat

import (
	"sync"
	"testing"
)

func TestMessage_Meta(t *testing.T) {
	// 1. Test initial state (should be nil)
	msg := &Message{}
	if meta := msg.Meta(); meta != nil {
		t.Errorf("Expected initial Meta() to return nil, got %v", meta)
	}

	// 2. Test after explicitly setting meta
	expectedMap := &sync.Map{}
	msg.meta = expectedMap // Directly set the private field for testing

	retrievedMap := msg.Meta()
	if retrievedMap != expectedMap {
		t.Errorf("Meta() did not return the expected sync.Map instance")
	}

	// 3. Verify we can use the returned map
	testKey := "testKey"
	testValue := "testValue"
	retrievedMap.Store(testKey, testValue)

	loadedValue, ok := msg.meta.Load(testKey) // Load directly from the internal field to confirm
	if !ok || loadedValue != testValue {
		t.Errorf("Value stored via map returned by Meta() was not found or incorrect in msg.meta")
	}

	loadedValueFromMethod, okFromMethod := retrievedMap.Load(testKey)
	if !okFromMethod || loadedValueFromMethod != testValue {
		t.Errorf("Value stored via map returned by Meta() was not retrievable via the same map instance")
	}
}

// Add tests for ContentString() if needed
// func TestMessage_ContentString(t *testing.T) { ... }
