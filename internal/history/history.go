// Package history は、チャット履歴の管理機能を提供します。
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// historyFile は、チャット履歴を保存するファイルの名前を定義します。
const historyFile = "chat_history.json"

// ChatMessage は、単一のチャットメッセージを表現する構造体です。
type ChatMessage struct {
	Role    string    `json:"role"`    // メッセージの送信者の役割（例：user, assistant）
	Content string    `json:"content"` // メッセージの内容
	Time    time.Time `json:"time"`    // メッセージが送信された時刻
}

// ChatHistory は、複数のChatMessageを含むチャット履歴を表現する構造体です。
type ChatHistory struct {
	Messages []ChatMessage `json:"messages"` // チャットメッセージのスライス
}

// AddMessage は、新しいメッセージをチャット履歴に追加します。
// role はメッセージの送信者の役割、content はメッセージの内容です。
func (h *ChatHistory) AddMessage(role, content string) {
	h.Messages = append(h.Messages, ChatMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
}

// LoadChatHistory は、ファイルからチャット履歴を読み込みます。
// 履歴ファイルが存在しない場合や読み込みエラーが発生した場合は、
// 新しい空のChatHistoryを返します。
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

// SaveChatHistory は、指定されたChatHistoryをJSONファイルに保存します。
// 保存に失敗した場合はエラーメッセージを出力します。
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
