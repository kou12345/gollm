package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type ChatMessage struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

type ChatHistory struct {
	Messages []ChatMessage `json:"messages"`
}

const historyFile = "chat_history.json"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")

	// Initialize the chat
	cs := model.StartChat()
	history := loadChatHistory()

	// Populate chat history
	for _, msg := range history.Messages {
		if msg.Role == "user" {
			cs.History = append(cs.History, &genai.Content{Role: "user", Parts: []genai.Part{genai.Text(msg.Content)}})
		} else {
			cs.History = append(cs.History, &genai.Content{Role: "model", Parts: []genai.Part{genai.Text(msg.Content)}})
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()

		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Exiting chat...")
			break
		}

		// Add user message to history
		history.Messages = append(history.Messages, ChatMessage{
			Role:    "user",
			Content: userInput,
			Time:    time.Now(),
		})

		resp, err := cs.SendMessage(ctx, genai.Text(userInput))
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			fmt.Print("Gemini: ")
			var responseContent string
			for _, part := range resp.Candidates[0].Content.Parts {
				responseContent += fmt.Sprint(part)
			}
			fmt.Println(responseContent)

			// Add Gemini's response to history
			history.Messages = append(history.Messages, ChatMessage{
				Role:    "assistant",
				Content: responseContent,
				Time:    time.Now(),
			})
		} else {
			fmt.Println("Gemini: No response received.")
		}
		fmt.Println()

		// Save chat history after each interaction
		saveChatHistory(history)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}

func loadChatHistory() ChatHistory {
	var history ChatHistory
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return ChatHistory{Messages: []ChatMessage{}}
		}
		log.Printf("Error reading chat history: %v\n", err)
		return ChatHistory{Messages: []ChatMessage{}}
	}

	err = json.Unmarshal(data, &history)
	if err != nil {
		log.Printf("Error unmarshaling chat history: %v\n", err)
		return ChatHistory{Messages: []ChatMessage{}}
	}

	return history
}

func saveChatHistory(history ChatHistory) {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		log.Printf("Error marshaling chat history: %v\n", err)
		return
	}

	err = os.WriteFile(historyFile, data, 0644)
	if err != nil {
		log.Printf("Error writing chat history: %v\n", err)
	}
}
