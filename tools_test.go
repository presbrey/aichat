package aichat_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

func TestContentParts(t *testing.T) {
	tests := []struct {
		name    string
		content any
		want    []*aichat.Part
		wantErr bool
	}{
		{
			name: "valid parts with text and image",
			content: []any{
				map[string]any{
					"type": "text",
					"text": "Hello world",
				},
				map[string]any{
					"type": "image_url",
					"image_url": map[string]any{
						"url":    "https://example.com/image.jpg",
						"detail": "high",
					},
				},
			},
			want: []*aichat.Part{
				{
					Type: "text",
					Text: "Hello world",
				},
				{
					Type: "image_url",
					ImageURL: struct {
						URL    string `json:"url"`
						Detail string `json:"detail,omitempty"`
					}{
						URL:    "https://example.com/image.jpg",
						Detail: "high",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "string content returns nil",
			content: "plain text",
			want:    nil,
			wantErr: false,
		},
		// {
		// 	name: "invalid part structure",
		// 	content: []any{
		// 		map[string]any{
		// 			"invalid": "structure",
		// 		},
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &aichat.Message{
				Content: tt.content,
			}
			got, err := msg.ContentParts()
			if tt.wantErr {
				assert.Error(t, err, "Expected error in ContentParts()")
			} else {
				assert.NoError(t, err, "Unexpected error in ContentParts()")
			}
			assert.Equal(t, tt.want, got, "ContentParts() result mismatch")
		})
	}
}

func TestArgumentsMap(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		want      map[string]interface{}
		wantError bool
	}{
		{
			name: "valid json",
			args: `{"location": "Boston, MA", "units": "celsius"}`,
			want: map[string]interface{}{
				"location": "Boston, MA",
				"units":    "celsius",
			},
			wantError: false,
		},
		{
			name:      "empty arguments",
			args:      "",
			want:      map[string]interface{}{},
			wantError: false,
		},
		{
			name:      "invalid json",
			args:      `{"location": "Boston, MA"`,
			want:      nil,
			wantError: true,
		},
		{
			name: "nested json",
			args: `{"location": {"city": "Boston", "state": "MA"}, "units": "celsius"}`,
			want: map[string]interface{}{
				"location": map[string]interface{}{
					"city":  "Boston",
					"state": "MA",
				},
				"units": "celsius",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &aichat.Function{Arguments: tt.args}
			got, err := fc.ArgumentsMap()

			if tt.wantError {
				assert.Error(t, err, "Expected error in ArgumentsMap()")
				return
			} else {
				assert.NoError(t, err, "Unexpected error in ArgumentsMap()")
			}
			assert.Equal(t, tt.want, got, "ArgumentsMap() result mismatch")
		})
	}
}

func TestRangePendingToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		messages []*aichat.Message
		wantIDs  []string
		wantErr  bool
	}{
		{
			name: "no pending tool calls",
			messages: []*aichat.Message{
				{Role: "assistant", ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
				{Role: "tool", ToolCallID: "call1", Content: "response1"},
			},
			wantIDs: nil,
			wantErr: false,
		},
		{
			name: "one pending tool call",
			messages: []*aichat.Message{
				{Role: "assistant", ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
			},
			wantIDs: []string{"call1"},
			wantErr: false,
		},
		{
			name: "multiple pending tool calls",
			messages: []*aichat.Message{
				{Role: "assistant", ToolCalls: []aichat.ToolCall{{ID: "call1"}, {ID: "call2"}}},
				{Role: "tool", ToolCallID: "call1", Content: "response1"},
			},
			wantIDs: []string{"call2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := &aichat.Chat{Messages: tt.messages}
			var gotIDs []string

			err := chat.RangePendingToolCalls(func(toolCall *aichat.ToolCallContext) error {
				gotIDs = append(gotIDs, toolCall.ToolCall.ID)
				return nil
			})

			if tt.wantErr {
				assert.Error(t, err, "Expected error in RangePendingToolCalls()")
				return
			} else {
				assert.NoError(t, err, "Unexpected error in RangePendingToolCalls()")
			}
			assert.Equal(t, tt.wantIDs, gotIDs, "Processed IDs do not match expected")
		})
	}

	t.Run("error case", func(t *testing.T) {
		chat := &aichat.Chat{
			Messages: []*aichat.Message{
				{Role: "assistant", ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
			},
		}

		expectedErr := "test error"
		err := chat.RangePendingToolCalls(func(toolCall *aichat.ToolCallContext) error {
			return errors.New(expectedErr)
		})

		assert.EqualError(t, err, expectedErr, "Expected specific error message in RangePendingToolCalls()")
	})
}

func TestToolCallContext(t *testing.T) {
	tests := []struct {
		name          string
		toolCall      *aichat.ToolCall
		wantName      string
		wantArgs      map[string]interface{}
		wantArgsError bool
		returnResult  map[string]interface{}
		wantError     bool
	}{
		{
			name: "valid tool call",
			toolCall: &aichat.ToolCall{
				ID:   "test1",
				Type: "function",
				Function: aichat.Function{
					Name:      "getWeather",
					Arguments: `{"location": "Boston", "units": "celsius"}`,
				},
			},
			wantName: "getWeather",
			wantArgs: map[string]interface{}{
				"location": "Boston",
				"units":    "celsius",
			},
			wantArgsError: false,
			returnResult: map[string]interface{}{
				"temperature": 20,
				"condition":   "sunny",
			},
			wantError: false,
		},
		{
			name: "invalid arguments json",
			toolCall: &aichat.ToolCall{
				ID:   "test2",
				Type: "function",
				Function: aichat.Function{
					Name:      "testFunc",
					Arguments: `{"invalid": json`,
				},
			},
			wantName:      "testFunc",
			wantArgs:      nil,
			wantArgsError: true,
			returnResult:  nil,
			wantError:     false,
		},
		{
			name: "unmarshalable result",
			toolCall: &aichat.ToolCall{
				ID:   "test3",
				Type: "function",
				Function: aichat.Function{
					Name:      "testFunc",
					Arguments: `{}`,
				},
			},
			wantName:      "testFunc",
			wantArgs:      map[string]interface{}{},
			wantArgsError: false,
			returnResult: map[string]interface{}{
				"channel": make(chan int), // channels cannot be marshaled to JSON
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := &aichat.Chat{}
			tcm := &aichat.ToolCallContext{
				ToolCall: tt.toolCall,
				Chat:     chat,
			}

			// Test Name()
			assert.Equal(t, tt.wantName, tcm.Name(), "ToolCallContext Name() mismatch")

			got, err := tcm.Arguments()
			if tt.wantArgsError {
				assert.Error(t, err, "Expected error in Arguments()")
				return
			} else {
				assert.NoError(t, err, "Unexpected error in Arguments()")
			}
			assert.Equal(t, tt.wantArgs, got, "Arguments() result mismatch")

			if tt.returnResult != nil {
				err := tcm.Return(tt.returnResult)
				if tt.wantError {
					assert.Error(t, err, "Expected error in Return()")
				} else {
					assert.NoError(t, err, "Unexpected error in Return()")
				}

				if !tt.wantError {
					assert.Equal(t, 1, len(chat.Messages), "Expected one message added to chat")
				}
			}
		})
	}
}
