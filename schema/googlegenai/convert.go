package googlegenai

import (
	"github.com/google/generative-ai-go/genai"
	"github.com/presbrey/aichat"
)

// ConvertTools converts the []*aichat.Tool to *genai.Tool
func ConvertTools(tools []*aichat.Tool) *genai.Tool {
	decls := make([]*genai.FunctionDeclaration, len(tools))
	for i, t := range tools {
		decls[i] = ToolToFunctionDeclaration(t)
	}
	return &genai.Tool{FunctionDeclarations: decls}
}

// stringToGenAIType converts a string type to genai.Type
func stringToGenAIType(t string) genai.Type {
	switch t {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}

// ToolToFunctionDeclaration converts an *aichat.Tool to a *genai.FunctionDeclaration
func ToolToFunctionDeclaration(t *aichat.Tool) *genai.FunctionDeclaration {
	// Convert our Parameters to genai.Schema
	schema := &genai.Schema{
		Type:       stringToGenAIType(t.Function.Parameters.Type),
		Properties: make(map[string]*genai.Schema),
		Required:   t.Function.Parameters.Required,
	}

	// Add properties from our parameter definitions
	for key, prop := range t.Function.Parameters.Properties {
		schema.Properties[key] = &genai.Schema{
			Type:        stringToGenAIType(prop.Type),
			Description: prop.Description,
		}
	}

	return &genai.FunctionDeclaration{
		Name:        t.Function.Name,
		Description: t.Function.Description,
		Parameters:  schema,
	}
}
