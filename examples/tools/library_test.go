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
	
	// Get all weather tools
	weatherTools := Get("weather")
	assert.NotEmpty(t, weatherTools, "Expected weather tools to be non-empty")
	
	// Check if get_weather_data exists in the returned tools
	found := false
	for _, tool := range weatherTools {
		if tool.Function.Name == "get_weather_data" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected 'get_weather_data' tool to be in the returned tools")
	
	assert.Nil(t, Get("nonexistent"), "Expected 'nonexistent' tool to be nil")
}
