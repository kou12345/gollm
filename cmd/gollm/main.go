package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kou12345/gollm/pkg/utils"

	_ "github.com/mattn/go-sqlite3"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChatRoom struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

// list.Item インターフェースを満たすためのメソッド
func (c ChatRoom) Title() string { return c.Name }
func (c ChatRoom) Description() string {
	return fmt.Sprintf("Created at: %s", c.CreatedAt.Format("2006-01-02 15:04:05"))
}
func (c ChatRoom) FilterValue() string { return c.Name }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				chatRoom := selectedItem.(ChatRoom)
				fmt.Printf("Selected chat room: %s (ID: %d)\n", chatRoom.Name, chatRoom.ID)
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(utils.ErrorColor("Error loading .env file"))
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal(utils.ErrorColor("GEMINI_API_KEY is not set in the environment"))
	}

	DbConnection, _ := sql.Open("sqlite3", "./db.sql")
	defer DbConnection.Close()

	cmd := `SELECT * FROM chat_rooms;`
	rows, err := DbConnection.Query(cmd)
	if err != nil {
		log.Fatalln(err)
	}

	var items []list.Item
	for rows.Next() {
		var chatRoom ChatRoom
		err = rows.Scan(&chatRoom.ID, &chatRoom.Name, &chatRoom.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, chatRoom)
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Chat Rooms"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// chat, err := chat.NewChat(option.WithAPIKey(apiKey))
// if err != nil {
// 	log.Fatal(utils.ErrorColor(err))
// }
// defer chat.Close()

// chat.Run()
