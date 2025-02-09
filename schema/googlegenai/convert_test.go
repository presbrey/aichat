package googlegenai

import (
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/presbrey/aichat"
	"github.com/stretchr/testify/assert"
)

func TestStringToGenAIType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected genai.Type
	}{
		{"string type", "string", genai.TypeString},
		{"number type", "number", genai.TypeNumber},
		{"integer type", "integer", genai.TypeInteger},
		{"boolean type", "boolean", genai.TypeBoolean},
		{"array type", "array", genai.TypeArray},
		{"object type", "object", genai.TypeObject},
		{"unknown type", "unknown", genai.TypeUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToGenAIType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToolToFunctionDeclaration(t *testing.T) {
	tool := &aichat.Tool{
		Function: aichat.Function{
			Name:        "testFunction",
			Description: "Test function description",
			Parameters: aichat.Parameters{
				Type:     "object",
				Required: []string{"param1"},
				Properties: map[string]aichat.Property{
					"param1": {
						Type:        "string",
						Description: "Parameter 1 description",
					},
					"param2": {
						Type:        "number",
						Description: "Parameter 2 description",
					},
				},
			},
		},
	}

	result := ToolToFunctionDeclaration(tool)

	assert.Equal(t, "testFunction", result.Name)
	assert.Equal(t, "Test function description", result.Description)
	assert.Equal(t, genai.TypeObject, result.Parameters.Type)
	assert.Equal(t, []string{"param1"}, result.Parameters.Required)
	assert.Equal(t, genai.TypeString, result.Parameters.Properties["param1"].Type)
	assert.Equal(t, "Parameter 1 description", result.Parameters.Properties["param1"].Description)
	assert.Equal(t, genai.TypeNumber, result.Parameters.Properties["param2"].Type)
	assert.Equal(t, "Parameter 2 description", result.Parameters.Properties["param2"].Description)
}

func TestConvertTools(t *testing.T) {
	tools := []*aichat.Tool{
		{
			Function: aichat.Function{
				Name:        "function1",
				Description: "Function 1 description",
				Parameters: aichat.Parameters{
					Type: "object",
					Properties: map[string]aichat.Property{
						"param1": {
							Type:        "string",
							Description: "Parameter 1",
						},
					},
				},
			},
		},
		{
			Function: aichat.Function{
				Name:        "function2",
				Description: "Function 2 description",
				Parameters: aichat.Parameters{
					Type: "object",
					Properties: map[string]aichat.Property{
						"param2": {
							Type:        "number",
							Description: "Parameter 2",
						},
					},
				},
			},
		},
	}

	result := ConvertTools(tools)

	assert.Len(t, result.FunctionDeclarations, 2)
	assert.Equal(t, "function1", result.FunctionDeclarations[0].Name)
	assert.Equal(t, "function2", result.FunctionDeclarations[1].Name)
}
