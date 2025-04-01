package aichat_test

import (
	"encoding/json"
	"testing"

	"github.com/presbrey/aichat"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Meta(t *testing.T) {
	// 1. Test initialization
	msg := &aichat.Message{}

	// Verify initial state is empty
	assert.Nil(t, msg.Meta().Get("foo"))

	// Verify storing and loading works on the initially returned map
	testKey1 := "testKey1"
	testValue1 := "testValue1"
	msg.Meta().Set(testKey1, testValue1)

	loadedValue1 := msg.Meta().Get(testKey1)
	assert.Equal(t, testValue1, loadedValue1)

	// 2. Test consistency: subsequent calls should return the same map values
	loadedValue1 = msg.Meta().Get(testKey1)
	assert.Equal(t, testValue1, loadedValue1)

	// Verify storing on the second map instance affects the first one (since they should be the same)
	testKey2 := "testKey2"
	testValue2 := "testValue2"
	msg.Meta().Set(testKey2, testValue2)

	loadedValue2 := msg.Meta().Get(testKey2)
	assert.Equal(t, testValue2, loadedValue2)

	// 3. Test keys
	keys := msg.Meta().Keys()
	assert.Equal(t, []string{testKey1, testKey2}, keys)

	// 4. Test Meta serialization
	jsonBytes, err := json.Marshal(msg.Meta())
	assert.NoError(t, err)
	assert.Equal(t, `{"testKey1":"testValue1","testKey2":"testValue2"}`, string(jsonBytes))
}

func TestMessage_Json(t *testing.T) {
	msg := &aichat.Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	jsonBytes, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Equal(t, `{"role":"user","content":"Hello, world!"}`, string(jsonBytes))
}
