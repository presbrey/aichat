package aichat_test

import (
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
