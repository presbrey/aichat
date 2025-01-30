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

## Installation

```bash
go get github.com/presbrey/aichat
```

## Usage

### Creating a New Chat

```go
import "github.com/presbrey/aichat"

// Initialize with S3 storage
s3Storage := YourS3Implementation{} // Implements aichat.S3 interface
options := aichat.Options{S3: s3Storage}

// Create new chat
chat := aichat.NewChat("chat-123", options)

// Or use storage wrapper
storage := aichat.NewChatStorage(options)
chat, err := storage.Load("chat-123")
```

### Managing Messages

```go
// Add messages
chat.AddUserMessage("Hello!")
chat.AddAssistantMessage("Hi! How can I help?")

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

// Add tool response
chat.AddToolResponse("get_weather", "call-123", `{"temp": 72, "condition": "sunny"}`)
```

### Working with Content

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

// Parse tool arguments
if toolCall := chat.Messages[2].ToolCalls[0]; toolCall.Type == "function" {
    args, err := toolCall.Function.ArgumentsMap()
    if err == nil {
        location := args["location"].(string)
        fmt.Println("Weather lookup for:", location)
    }
}
```

### Storage Operations

```go
// Save chat state
err := chat.Save("chat-123")

// Load existing chat
err = chat.Load("chat-123")

// Delete chat
err = chat.Delete("chat-123")
```

### Direct Access and Helper Methods

The `Chat` and `Message` structs are designed to be transparent - you are welcome to access their members directly in your applications. For example, you can directly access `chat.Messages`, `chat.Meta`, or `message.Role`.

For convenience, the package also provides several helper methods:

- `AddRoleContent(role, content)`: Add a message with any role and content
- `AddUserContent(content)`: Add a user message
- `AddAssistantContent(content)`: Add an assistant message
- `AddToolRawContent(name, toolCallID, content)`: Add a tool message with raw content
- `AddToolContent(name, toolCallID, content)`: Add a tool message with JSON-encoded content if needed
- `AddAssistantToolCall(toolCalls)`: Add an assistant message with tool calls
- `LastMessage()`: Get the most recent message
- `LastMessageRole()`: Get the role of the most recent message
- `Range(fn)`: Iterate through messages with a callback function

```go
// Example of direct member access
fmt.Println(chat.ID, chat.LastUpdated)
for _, msg := range chat.Messages {
    fmt.Println(msg.Role, msg.Content)
}

// Example of helper method usage
chat.AddUserContent("Hello")
if last := chat.LastMessage(); last != nil {
    fmt.Println("Last message was from:", last.Role)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

Copyright (c) 2025 Joe Presbrey
