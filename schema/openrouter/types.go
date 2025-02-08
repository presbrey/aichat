package openrouter

import "github.com/presbrey/aichat"

// Request represents a chat completion request to the OpenRouter API
type Request struct {
	Messages       []*aichat.Message `json:"messages,omitempty"`
	Prompt         string            `json:"prompt,omitempty"`
	Model          string            `json:"model,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`

	Stop        interface{} `json:"stop,omitempty"` // string or []string
	Stream      bool        `json:"stream,omitempty"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Temperature float64     `json:"temperature,omitempty"`

	Tools      []*aichat.Tool `json:"tools,omitempty"`
	ToolChoice interface{}    `json:"tool_choice,omitempty"` // string or ToolChoice

	Seed        int     `json:"seed,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	FreqPenalty float64 `json:"frequency_penalty,omitempty"`
	PresPenalty float64 `json:"presence_penalty,omitempty"`
	RepPenalty  float64 `json:"repetition_penalty,omitempty"`

	LogitBias   map[int]float64 `json:"logit_bias,omitempty"`
	TopLogprobs int             `json:"top_logprobs,omitempty"`

	MinP float64 `json:"min_p,omitempty"`
	TopA float64 `json:"top_a,omitempty"`

	Prediction *struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	} `json:"prediction,omitempty"`
	Transforms []string `json:"transforms,omitempty"`
	Models     []string `json:"models,omitempty"`
	Route      string   `json:"route,omitempty"`

	IncludeReasoning bool `json:"include_reasoning,omitempty"`
}

// Response represents the API response structure
type Response struct {
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`

	UserID   string `json:"user_id,omitempty"`
	ID       string `json:"id,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	Object   string `json:"object,omitempty"`
	Created  int64  `json:"created,omitempty"`

	SystemFingerprint string `json:"system_fingerprint,omitempty"`

	Choices []struct {
		LogProbs           interface{}     `json:"logprobs"`
		FinishReason       string          `json:"finish_reason"`
		NativeFinishReason string          `json:"native_finish_reason"`
		Index              int             `json:"index"`
		Message            *aichat.Message `json:"message"`
	} `json:"choices,omitempty"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// ToolChoice represents the model's choice of tool usage
type ToolChoice struct {
	Type string `json:"type,omitempty"`

	Function struct {
		Name string `json:"name"`
	} `json:"function,omitempty"`
}
