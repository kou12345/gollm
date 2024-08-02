package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

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
	cs.History = []*genai.Content{}

	// Create a scanner to read user input
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()

		// Check if user wants to exit
		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Exiting chat...")
			break
		}

		resp, err := cs.SendMessage(ctx, genai.Text(userInput))
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			fmt.Print("Gemini: ")
			for _, part := range resp.Candidates[0].Content.Parts {
				fmt.Print(part)
			}
			fmt.Println()
		} else {
			fmt.Println("Gemini: No response received.")
		}
		fmt.Println() // Add a blank line for readability
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}
