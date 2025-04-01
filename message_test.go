package aichat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage_Meta(t *testing.T) {
	// 1. Test initialization
	msg := &Message{}

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
}
