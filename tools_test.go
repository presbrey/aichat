package aichat_test

import (
	"errors"
	"reflect"
	"testing"

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
			if (err != nil) != tt.wantErr {
				t.Errorf("ContentParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContentParts() = %v, want %v", got, tt.want)
			}
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

			if (err != nil) != tt.wantError {
				t.Errorf("ArgumentsMap() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArgumentsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangePendingToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		messages []aichat.Message
		wantIDs  []string
		wantErr  bool
	}{
		{
			name: "no pending tool calls",
			messages: []aichat.Message{
				{ToolCallID: "call1", Content: "response1"},
				{ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
			},
			wantIDs: nil,
			wantErr: false,
		},
		{
			name: "one pending tool call",
			messages: []aichat.Message{
				{ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
			},
			wantIDs: []string{"call1"},
			wantErr: false,
		},
		{
			name: "multiple pending tool calls",
			messages: []aichat.Message{
				{ToolCalls: []aichat.ToolCall{{ID: "call1"}, {ID: "call2"}}},
				{ToolCallID: "call1", Content: "response1"},
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

			if (err != nil) != tt.wantErr {
				t.Errorf("RangePendingToolCalls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("RangePendingToolCalls() processed IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}

	t.Run("error case", func(t *testing.T) {
		chat := &aichat.Chat{
			Messages: []aichat.Message{
				{ToolCalls: []aichat.ToolCall{{ID: "call1"}}},
			},
		}

		expectedErr := "test error"
		err := chat.RangePendingToolCalls(func(toolCall *aichat.ToolCallContext) error {
			return errors.New(expectedErr)
		})

		if err == nil || err.Error() != expectedErr {
			t.Errorf("RangePendingToolCalls() error = %v, want %v", err, expectedErr)
		}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := &aichat.Chat{}
			tcm := &aichat.ToolCallContext{
				ToolCall: tt.toolCall,
				Chat:     chat,
			}

			// Test Name()
			if got := tcm.Name(); got != tt.wantName {
				t.Errorf("Name() = %v, want %v", got, tt.wantName)
			}

			// Test Arguments()
			got, err := tcm.Arguments()
			if (err != nil) != tt.wantArgsError {
				t.Errorf("Arguments() error = %v, wantArgsError %v", err, tt.wantArgsError)
				return
			}
			if !tt.wantArgsError && !reflect.DeepEqual(got, tt.wantArgs) {
				t.Errorf("Arguments() = %v, want %v", got, tt.wantArgs)
			}

			// Test Return()
			if tt.returnResult != nil {
				err := tcm.Return(tt.returnResult)
				if (err != nil) != tt.wantError {
					t.Errorf("Return() error = %v, wantError %v", err, tt.wantError)
				}

				// Verify the response was added to the chat
				if !tt.wantError && len(chat.Messages) != 1 {
					t.Error("Return() did not add message to chat")
				}
			}
		})
	}
}
