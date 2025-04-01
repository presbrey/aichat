# AI Chat Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/presbrey/aichat)](https://goreportcard.com/report/github.com/presbrey/aichat)
[![codecov](https://codecov.io/gh/presbrey/aichat/graph/badge.svg?token=PHVQ7QN4TL)](https://codecov.io/gh/presbrey/aichat)
[![Go](https://github.com/presbrey/aichat/actions/workflows/go.yml/badge.svg)](https://github.com/presbrey/aichat/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/presbrey/aichat.svg)](https://pkg.go.dev/github.com/presbrey/aichat)

Simple Go package for managing AI chat sessions across LLM Providers with options for message history, tool calling, and S3-compatible session storage. Works with OpenRouter, OpenAI, Google GenAI, DeepSeek, and many others with a Chat Completion API.

The [toolcalling example](examples/toolcalling/example_test.go) uses [OpenRouter API](https://openrouter.ai/docs/api-reference/overview) with GPT-4o and [OpenAI Function schema](https://platform.openai.com/docs/guides/function-calling). The tool converter in the [googlegenai](schema/googlegenai/convert.go) subpackage provides support for [Google GenAI SDK](https://github.com/google/generative-ai-go). Define tools once [in YAML](examples/tools/library.yaml) or JSON and reuse them across sessions, providers, and SDKs.

## Features

- Chat session management with message history and timestamps
- Support for multiple message types (system, user, assistant, tool)
- Comprehensive message operations:
  - Add, remove, pop, shift, and unshift messages
  - Query by role and content type
  - Message counting and iteration
  - Idempotent message handling
- Tool/Function calling system:
  - Structured tool calls with ID tracking
  - JSON argument parsing
  - Pending tool call management
  - Tool response handling
- S3-compatible storage backend:
  - Load/Save/Delete operations
  - Context-aware storage operations
  - Support for any S3-compatible service
- Rich content support:
  - Text content
  - Structured content (JSON)
  - Multi-part content handling
- Session metadata:
  - Unique session IDs
  - Creation and update timestamps
  - Custom metadata storage
- JSON serialization with custom marshaling

## Usage

The `Chat`, `Message`, and `ToolCall` structs are designed to be transparent - applications are welcome to access their members directly. For example, you can directly access `chat.Messages`, `chat.Meta`, or `message.Role`.

For convenience, the package provides several helper methods:

### Chat Methods

- `AddMessage(msg *Message)`: Add a Message to the chat
- `AddMessageOnce(msg *Message)`: Add a Message to the chat (idempotent)
- `AddRoleContent(role string, content any) *Message`: Add a message with any role and content, returns the created message
- `AddUserContent(content any) *Message`: Add a user message, returns the created message
- `AddAssistantContent(content any) *Message`: Add an assistant message, returns the created message
- `AddToolRawContent(name string, toolCallID string, content any) *Message`: Add a tool message with raw content, returns the created message
- `AddToolContent(name string, toolCallID string, content any) error`: Add a tool message with JSON-encoded content if needed, returns error if JSON marshaling fails
- `AddAssistantToolCall(toolCalls []ToolCall) *Message`: Add an assistant message with tool calls, returns the created message
- `ClearMessages()`: Remove all messages from the chat
- `LastMessage() *Message`: Get the most recent message
- `LastMessageRole() string`: Get the role of the most recent message
- `LastMessageByRole(role string) *Message`: Get the last message with a specific role
- `LastMessageByType(contentType string) *Message`: Get the last message with a specific content type
- `MessageCount() int`: Get the total number of messages in the chat
- `MessageCountByRole(role string) int`: Get the count of messages with a specific role
- `PopMessage() *Message`: Remove and return the last message from the chat
- `PopMessageIfRole(role string) *Message`: Remove and return the last message if it matches the specified role
- `Range(fn func(msg *Message) error) error`: Iterate through messages with a callback function
- `RangeByRole(role string, fn func(msg *Message) error) error`: Iterate through messages with a specific role
- `RemoveLastMessage() *Message`: Remove and return the last message from the chat (alias for PopMessage)
- `SetSystemContent(content any) *Message`: Set or update the system message content at the beginning of the chat, returns the system message
- `SetSystemMessage(msg *Message) *Message`: Set or update the system message at the beginning of the chat, returns the system message
- `ShiftMessages() *Message`: Remove and return the first message from the chat
- `UnshiftMessages(msg *Message)`: Insert a message at the beginning of the chat

### Message Methods

- `Meta() *Meta`: Get a Meta struct for working with message metadata
- `ContentString() string`: Get the content as a string if it's a simple string
- `ContentParts() ([]*Part, error)`: Get the content as a slice of Part structs if it's a multipart message

### Meta Methods

- `Set(key string, value any)`: Set a metadata value on a Message
- `Get(key string) any`: Retrieve a metadata value from a Message
- `Keys() []string`: Get all metadata keys for a Message

### Function Methods

- `ArgumentsMap() map[string]any`: Parse and return a map from a Function's Arguments JSON

### Creating a New Chat

```go
// Create new chat in-memory
chat := new(aichat.Chat)

// Or use persistent/S3-compatible storage wrapper
opts := aichat.Options{...}
storage := aichat.NewChatStorage(opts)
chat, err := storage.Load("chat-f00ba0ba0")
```

### Working with Messages

The `[]*Message` structure can be used to manage messages in a chat session in multiple ways:

```go
// Add a message directly (idempotent)
chat.AddMessage(&aichat.Message{
    Role: "user",
    Content: "Hello!",
})

// Add user content (creates new message)
chat.AddUserContent("Hello!")

// Add assistant content (creates new message)
chat.AddAssistantContent("Hi there!")

// Set or update the system message
chat.SetSystemContent("Welcome to the chat!")

// Remove the last message if it's from the assistant
if msg := chat.PopMessageIfRole("assistant"); msg != nil {
    fmt.Println("Removed assistant's last message:", msg.Content)
}

// Get the last message
if last := chat.LastMessage(); last != nil {
    fmt.Println("Last message was from:", last.Role) // "assistant"
}

// Example of direct member access
fmt.Println(chat.ID, chat.LastUpdated)
for _, msg := range chat.Messages {
    fmt.Println(msg.Role, msg.Content)
}

// Add tool/function calls
toolCalls := []aichat.ToolCall{{
    ID:   "call-123",
    Type: "function",
        Function: aichat.Function{
        Name:      "get_weather",
        Arguments: `{"location": "Boston"}`,
        },
    },
}
chat.AddAssistantToolCall(toolCalls)
```

### Working with Strings and Multi-Part Content

The `Message` struct provides a `ContentString()` method that returns the content as a string if it is a simple string.
The `ContentParts()` method returns the content as a slice of `Part` structs if it is a multipart message.

```go
// Handle text content
textMsg := message.ContentString()
fmt.Printf("Text: %s\n", textMsg)

// Handle rich content (text/images)
if parts, err := message.ContentParts(); err == nil {
    for _, part := range parts {
        switch part.Type {
        case "text":
            fmt.Println("Text:", part.Text)
        case "image_url":
            fmt.Println("Image:", part.ImageURL.URL)
        }
    }
}

// Working with message metadata
message.Meta().Set("timestamp", time.Now())
message.Meta().Set("processed", true)

timestamp := message.Meta().Get("timestamp")
keys := message.Meta().Keys() // Get all metadata keys
```

### Handling Pending Tool Calls

```go
// Iterate over pending tool calls
err := chat.RangePendingToolCalls(func(tcc *aichat.ToolCallContext) error {
    // Get the name of the tool/function
    name := tcc.Name()

    // Get the arguments of the tool call
    args, err := tcc.Arguments()
    if err != nil {
        return err
    }

    // Handle the tool call based on its name
    switch name {
    case "get_weather":
        // Implement the logic for the "get_weather" tool
        location, _ := args["location"].(string)
        weatherData := getWeatherData(location) // Replace with your implementation

        // Return the result back to the chat session
        return tcc.Return(map[string]any{
            "location": location,
            "weather":  weatherData,
        })
    default:
        return fmt.Errorf("unknown tool: %s", name)
    }
})

if err != nil {
    fmt.Println("Error processing tool calls:", err)
}
```

### Chat Persistence via S3 Interface

The `Chat` struct provides methods for saving, loading, and deleting chat sessions. Pass a key (string) that will be used to lookup the chat in the storage backend. The `S3` interface is used to abstract the storage backend. Official AWS S3, Minio, Tigris, and others are compatible.

```go
userSessionKey := "user-123-chat-789"

// Save chat state
err := chat.Save(userSessionKey)

// Load existing chat
err = chat.Load(userSessionKey)

// Delete chat
err = chat.Delete(userSessionKey)

// Your S3 storage implementation should satisfy this interface:
type S3 interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, data io.Reader) error
	Delete(ctx context.Context, key string) error
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

Copyright (c) 2025 Joe Presbrey
