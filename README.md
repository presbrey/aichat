# AI Chat Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/presbrey/aichat)](https://goreportcard.com/report/github.com/presbrey/aichat)
[![codecov](https://codecov.io/gh/presbrey/aichat/graph/badge.svg?token=PHVQ7QN4TL)](https://codecov.io/gh/presbrey/aichat)
[![Go](https://github.com/presbrey/aichat/actions/workflows/go.yml/badge.svg)](https://github.com/presbrey/aichat/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/presbrey/aichat.svg)](https://pkg.go.dev/github.com/presbrey/aichat)

A Go package for managing AI chat sessions with support for message history, tool calls, and S3-compatible storage.

## Features

- Chat session management with message history
- Support for multiple message types (user, assistant, tool)
- Function calling and tool execution
- S3-compatible storage backend
- Rich content support including text and images
- Metadata storage for sessions
- JSON serialization

## Usage

### Creating a New Chat

```go
// Create new chat in-memory
chat := new(aichat.Chat)

// Or use persistent/S3-compatible storage wrapper
opts := aichat.Options{...}
storage := aichat.NewChatStorage(opts)
chat, err := storage.Load("chat-f00ba0ba0")
```

### Convinence Methods and Direct Access

The `Chat`, `Message`, and `ToolCall` structs are designed to be transparent - you are welcome to access their members directly in your applications. For example, you can directly access `chat.Messages`, `chat.Meta`, or `message.Role`.

For convenience, the package also provides several helper methods:

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
- `Range(fn)`: Iterate through messages with a callback function
- `RangeByRole(role, fn)`: Iterate through messages with a specific role
- `RemoveLastMessage()`: Remove and return the last message from the chat

```go
// Example of helper method usage
chat.AddUserContent("Hello")
chat.AddAssistantContent("Hi! How can I help?")
if last := chat.LastMessage(); last != nil {
    fmt.Println("Last message was from:", last.Role) // "assistant"
}

// Example of direct member access
fmt.Println(chat.ID, chat.LastUpdated)
for _, msg := range chat.Messages {
    fmt.Println(msg.Role, msg.Content)
}

// Add tool/function calls
toolCalls := []aichat.ToolCall{
    {
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
textMsg := chat.Messages[0].ContentString()

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

### Storage Operations

The `Chat` struct provides methods for saving, loading, and deleting chat sessions. Pass a key (string) that will be used to lookup the chat in the storage backend.

```go
// Save chat state
err := chat.Save("chat-123")

// Load existing chat
err = chat.Load("chat-123")

// Delete chat
err = chat.Delete("chat-123")
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

Copyright (c) 2025 Joe Presbrey
