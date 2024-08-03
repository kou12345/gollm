package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const historyFile = "chat_history.json"

type ChatMessage struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type ChatHistory struct {
	Messages []ChatMessage `json:"messages"`
}

func (h *ChatHistory) AddMessage(role, content string) {
	h.Messages = append(h.Messages, ChatMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
}

func LoadChatHistory() *ChatHistory {
	var history ChatHistory
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No existing chat history found. Starting a new conversation.")
			return &ChatHistory{Messages: []ChatMessage{}}
		}
		fmt.Printf("Failed to read chat history file: %v\nStarting with an empty history.\n", err)
		return &ChatHistory{Messages: []ChatMessage{}}
	}

	err = json.Unmarshal(data, &history)
	if err != nil {
		fmt.Printf("Error occurred while parsing chat history: %v\nStarting with an empty history.\n", err)
		return &ChatHistory{Messages: []ChatMessage{}}
	}

	fmt.Printf("Successfully loaded chat history with %d messages.\n", len(history.Messages))
	return &history
}

func SaveChatHistory(history ChatHistory) {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		fmt.Printf("Failed to convert chat history to JSON: %v\nChat history not saved.\n", err)
		return
	}

	err = os.WriteFile(historyFile, data, 0644)
	if err != nil {
		fmt.Printf("Failed to save chat history to file: %v\nPlease check file permissions or disk space.\n", err)
	} else {
		fmt.Println("Chat history successfully saved.")
	}
}
