package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolsLibrary(t *testing.T) {
	assert.NotEmpty(t, Library, "Expected non-empty library")
	assert.NotEmpty(t, Library["weather"], "Expected 'weather' tool to be defined")
	assert.NotEmpty(t, Library["weather"]["get_weather_data"], "Expected 'get_weather_data' tool to be defined")
	assert.Equal(t, "get_weather_data", Library["weather"]["get_weather_data"].Function.Name)
	assert.Equal(t, "get_weather_data", Get("weather")[0].Function.Name)
	assert.Nil(t, Get("nonexistent"), "Expected 'nonexistent' tool to be nil")
}
