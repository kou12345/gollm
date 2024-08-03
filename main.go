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

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
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

// TODO チャットの履歴一覧を表示する
// TODO チャットの履歴一覧から選択して再開する
// TODO チャットの履歴を削除する
// TODO Markdownで表示することとstreamingの両立

var (
	errorColor   = color.New(color.FgRed).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	userColor    = color.New(color.FgCyan).SprintFunc()
	aiColor      = color.New(color.FgYellow).SprintFunc()
)

func renderMarkdown(md string) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	out, err := r.Render(md)
	if err != nil {
		return md // エラーが発生した場合は元のテキストを返す
	}
	return out
}

func loadChatHistory() ChatHistory {
	var history ChatHistory
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No existing chat history found. Starting a new conversation.")
			return ChatHistory{Messages: []ChatMessage{}}
		}
		fmt.Printf("Failed to read chat history file: %v\nStarting with an empty history.\n", err)
		return ChatHistory{Messages: []ChatMessage{}}
	}

	err = json.Unmarshal(data, &history)
	if err != nil {
		fmt.Printf("Error occurred while parsing chat history: %v\nStarting with an empty history.\n", err)
		return ChatHistory{Messages: []ChatMessage{}}
	}

	fmt.Printf("Successfully loaded chat history with %d messages.\n", len(history.Messages))
	return history
}

func saveChatHistory(history ChatHistory) {
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(errorColor("Error loading .env file"))
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(errorColor(err))
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

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
		fmt.Print(userColor("You: "))
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()

		if strings.ToLower(userInput) == "exit" {
			fmt.Println(successColor("Exiting chat..."))
			break
		}

		// Add user message to history
		history.Messages = append(history.Messages, ChatMessage{
			Role:    "user",
			Content: userInput,
			Time:    time.Now(),
		})

		prompt := genai.Text(userInput)
		iter := cs.SendMessageStream(ctx, prompt)

		var fullResponse string
		fmt.Print(aiColor("Gemini: "))

		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Println(errorColor(fmt.Sprintf("Error occurred while receiving response: %v", err)))
				break
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				for _, part := range resp.Candidates[0].Content.Parts {
					partContent := fmt.Sprint(part)
					fullResponse += partContent
					// ストリーミング中の出力を削除
					// fmt.Print(partContent)
				}
			}
		}

		fmt.Println() // 改行を追加

		if fullResponse != "" {
			// レンダリングされたマークダウンを表示
			fmt.Print(aiColor("Gemini: "))
			fmt.Println(renderMarkdown(fullResponse))

			// Add Gemini's complete response to history
			history.Messages = append(history.Messages, ChatMessage{
				Role:    "assistant",
				Content: fullResponse,
				Time:    time.Now(),
			})

			// Save chat history after the complete response is received
			saveChatHistory(history)
		} else {
			fmt.Println(errorColor("Gemini: No response received. The AI model might be experiencing issues."))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(errorColor(fmt.Sprintf("Error occurred while reading input: %v\nExiting program.", err)))
	}
}
