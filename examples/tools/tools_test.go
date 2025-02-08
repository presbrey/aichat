package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibrary(t *testing.T) {
	assert.NotEmpty(t, Library, "Expected non-empty library")
	assert.NotEmpty(t, Library["weather"], "Expected 'weather' tool to be defined")
	assert.NotEmpty(t, Library["weather"][0], "Expected 'weather' tool to be defined")
	assert.Equal(t, "get_weather_data", Library["weather"][0].Function.Name)
}
