package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/kou12345/gollm/internal/history"
	"github.com/kou12345/gollm/internal/render"
	"github.com/kou12345/gollm/pkg/utils"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Chat struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	cs      *genai.ChatSession
	history *history.ChatHistory
	scanner *bufio.Scanner
}

func NewChat(opts ...option.ClientOption) (*Chat, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	cs := model.StartChat()
	history := history.LoadChatHistory()

	for _, msg := range history.Messages {
		if msg.Role == "user" {
			cs.History = append(cs.History, &genai.Content{Role: "user", Parts: []genai.Part{genai.Text(msg.Content)}})
		} else {
			cs.History = append(cs.History, &genai.Content{Role: "model", Parts: []genai.Part{genai.Text(msg.Content)}})
		}
	}

	return &Chat{
		client:  client,
		model:   model,
		cs:      cs,
		history: history,
		scanner: bufio.NewScanner(os.Stdin),
	}, nil
}

func (c *Chat) Close() {
	c.client.Close()
}

func (c *Chat) Run() {
	for {
		fmt.Print(utils.UserColor("You: "))
		if !c.scanner.Scan() {
			break
		}
		userInput := c.scanner.Text()

		if strings.ToLower(userInput) == "exit" {
			fmt.Println(utils.SuccessColor("Exiting chat..."))
			break
		}

		c.history.AddMessage("user", userInput)

		response := c.sendMessage(userInput)

		if response != "" {
			fmt.Print(utils.AIColor("Gemini: "))
			fmt.Println(render.RenderMarkdown(response))

			c.history.AddMessage("assistant", response)
			history.SaveChatHistory(*c.history)
		} else {
			fmt.Println(utils.ErrorColor("Gemini: No response received. The AI model might be experiencing issues."))
		}
	}

	if err := c.scanner.Err(); err != nil {
		fmt.Println(utils.ErrorColor(fmt.Sprintf("Error occurred while reading input: %v\nExiting program.", err)))
	}
}

func (c *Chat) sendMessage(message string) string {
	ctx := context.Background()
	prompt := genai.Text(message)
	iter := c.cs.SendMessageStream(ctx, prompt)

	var fullResponse string
	fmt.Print(utils.AIColor("Gemini: "))

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(utils.ErrorColor(fmt.Sprintf("Error occurred while receiving response: %v", err)))
			break
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				partContent := fmt.Sprint(part)
				fullResponse += partContent
				// Uncomment the following line to enable streaming output
				// fmt.Print(partContent)
			}
		}
	}

	fmt.Println()
	return fullResponse
}
