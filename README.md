# AI Chat Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/presbrey/aichat)](https://goreportcard.com/report/github.com/presbrey/aichat)
[![codecov](https://codecov.io/gh/presbrey/aichat/graph/badge.svg?token=PHVQ7QN4TL)](https://codecov.io/gh/presbrey/aichat)
[![Go](https://github.com/presbrey/aichat/actions/workflows/go.yml/badge.svg)](https://github.com/presbrey/aichat/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/presbrey/aichat.svg)](https://pkg.go.dev/github.com/presbrey/aichat)

Simple Go package for managing AI chat sessions across all LLM Providers with options for message history, tool calling, and S3-compatible session storage. Works with OpenRouter, OpenAI, Google GenAI, and many others.

The [toolcalling example](examples) uses [OpenRouter API](https://openrouter.ai/docs/api-reference/overview) with GPT-4o and [OpenAI Function schema](https://platform.openai.com/docs/guides/function-calling). The tool converter in the [googlegenai](schema/googlegenai) subpackage provides support for [Google GenAI SDK](https://github.com/google/generative-ai-go). Define tools once [in YAML](examples/tools/tools.yaml) or JSON and reuse them across sessions and SDKs.

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

For convenience, the package also provides several helper methods:

- `AddMessage(msg)`: Add a Message
- `AddRoleContent(role, content)`: Add a message with any role and content
- `AddUserContent(content)`: Add a user message
- `AddAssistantContent(content)`: Add an assistant message
- `AddToolRawContent(name, toolCallID, content)`: Add a tool message with raw content
- `AddToolContent(name, toolCallID, content)`: Add a tool message with JSON-encoded content if needed
- `AddAssistantToolCall(toolCalls)`: Add an assistant message with tool calls
- `ClearMessages()`: Remove all messages from the chat
- `LastMessage()`: Get the most recent message
- `LastMessageRole()`: Get the role of the most recent message
- `LastMessageByRole(role)`: Get the last message with a specific role
- `LastMessageByType(contentType)`: Get the last message with a specific content type
- `MessageCount()`: Get the total number of messages in the chat
- `MessageCountByRole(role)`: Get the count of messages with a specific role
- `PopMessage()`: Remove and return the last message from the chat
- `PopMessageIfRole(role)`: Remove and return the last message if it matches the specified role
- `Range(fn)`: Iterate through messages with a callback function
- `RangeByRole(role, fn)`: Iterate through messages with a specific role
- `RangePendingToolCalls(fn)`: Iterate over pending tool calls with a callback function
- `SetSystemContent(content)`: Set or update the system message content at the beginning of the chat
- `SetSystemMessage(msg)`: Set or update the system message at the beginning of the chat
- `ShiftMessages()`: Remove and return the first message from the chat
- `UnshiftMessages(msg)`: Insert a message at the beginning of the chat

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
