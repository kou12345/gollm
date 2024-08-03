package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kou12345/gollm/internal/chat"
	"github.com/kou12345/gollm/pkg/utils"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(utils.ErrorColor("Error loading .env file"))
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal(utils.ErrorColor("GEMINI_API_KEY is not set in the environment"))
	}

	chat, err := chat.NewChat(option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(utils.ErrorColor(err))
	}
	defer chat.Close()

	chat.Run()
}
