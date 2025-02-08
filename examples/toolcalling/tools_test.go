package toolcalling

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/presbrey/aichat"
	"github.com/presbrey/aichat/examples/tools"
	"github.com/presbrey/aichat/openrouter"
	"github.com/stretchr/testify/assert"
)

func TestToolCallingExample(t *testing.T) {
	mockServer := mockOpenRouter(t)
	defer mockServer.Close()
	llmURL := mockServer.URL
	// llmURL := "https://openrouter.ai/api/v1/chat/completions"

	newChat := new(aichat.Chat)
	newChat.AddMessage(&aichat.Message{
		Role:    "user",
		Content: "What is the weather in New York City on May 25th?",
	})

	req := &openrouter.Request{
		Messages: newChat.Messages,
		Model:    "openai/gpt-4o-2024-11-20",

		Tools:      tools.Library["weather"],
		ToolChoice: "auto",
	}

	jsonData, err := json.Marshal(req)
	assert.NoError(t, err)

	httpReq, err := http.NewRequest("POST", llmURL, bytes.NewBuffer(jsonData))
	assert.NoError(t, err)

	httpReq.Header.Set("Authorization", "Bearer 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")
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var result openrouter.Response
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	newChat.AddMessage(result.Choices[0].Message)

	newChat.RangePendingToolCalls(func(tc *aichat.ToolCallContext) error {
		assert.Equal(t, "get_weather_data", tc.Name())
		args, err := tc.Arguments()
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"datetime": "2023-05-25",
			"location": "New York City",
		}, args)
		return tc.Return(map[string]any{
			"temperature": 20,
			"condition":   "sunny",
		})
	})

	assert.Equal(t, 3, len(newChat.Messages))
	assert.Equal(t, "{\"condition\":\"sunny\",\"temperature\":20}", newChat.LastMessage().Content)
	assert.Equal(t, "What is the weather in New York City on May 25th?", newChat.LastMessageByRole("user").Content)
}

func mockOpenRouter(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
  "id": "gen-1739050836-Gzy8ygu6L6Dows3Y12pi",
  "provider": "OpenAI",
  "model": "openai/gpt-4o-2024-11-20",
  "object": "chat.completion",
  "created": 1739050836,
  "system_fingerprint": "fp_e53e529665",
  "choices": [
    {
      "logprobs": null,
      "finish_reason": "tool_calls",
      "native_finish_reason": "tool_calls",
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "",
        "tool_calls": [
          {
            "id": "call_DEg41BT7HSKP7cKzXGq7WFyI",
            "type": "function",
            "function": {
              "name": "get_weather_data",
              "arguments": "{\"location\":\"New York City\", \"datetime\":\"2023-05-25\"}"
            }
          }
        ]
      }
    }
  ],
  "usage": {
    "prompt_tokens": 107,
    "completion_tokens": 22,
    "total_tokens": 129
  }
}`))

		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
