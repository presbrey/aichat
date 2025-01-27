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

			err := chat.RangePendingToolCalls(func(toolCall *aichat.ToolCallMessage) error {
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
		err := chat.RangePendingToolCalls(func(toolCall *aichat.ToolCallMessage) error {
			return errors.New(expectedErr)
		})

		if err == nil || err.Error() != expectedErr {
			t.Errorf("RangePendingToolCalls() error = %v, want %v", err, expectedErr)
		}
	})
}
