package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/presbrey/aichat"
)

type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []*aichat.Message `json:"messages"`
	Stream   bool              `json:"stream"`
}

type DeepSeekResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index   int             `json:"index"`
	Message *aichat.Message `json:"message"`
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("DEEPSEEK_API_KEY environment variable is required")
		return
	}

	// Create a new chat session
	chat := &aichat.Chat{}

	// Add system and user messages
	chat.SetSystemContent("You are a helpful assistant.")
	chat.AddUserContent("Hello!")

	// Prepare the request
	reqBody := DeepSeekRequest{
		Model:    "deepseek-chat",
		Messages: chat.Messages,
		Stream:   false,
	}

	// Marshal request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Parse response
	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// Add assistant's response to chat
	if len(deepseekResp.Choices) > 0 {
		chat.AddMessage(deepseekResp.Choices[0].Message)
		fmt.Println("Assistant:", deepseekResp.Choices[0].Message.ContentString())
	}
}
