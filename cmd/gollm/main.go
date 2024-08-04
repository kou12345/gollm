package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/joho/godotenv"
	"github.com/kou12345/gollm/internal/render"
	"github.com/kou12345/gollm/pkg/utils"

	_ "github.com/mattn/go-sqlite3"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type ChatRoom struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

func (c ChatRoom) Title() string { return c.Name }
func (c ChatRoom) Description() string {
	return fmt.Sprintf("Created at: %s", c.CreatedAt.Format("2006-01-02 15:04:05"))
}
func (c ChatRoom) FilterValue() string { return c.Name }

type Message struct {
	ID         int
	ChatRoomID int
	Content    string
	CreatedAt  time.Time
}

type model struct {
	list         list.Model
	viewport     viewport.Model
	selectedRoom *ChatRoom
	messages     []Message
	state        string
	db           *sql.DB
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "list":
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				selectedItem := m.list.SelectedItem()
				if selectedItem != nil {
					m.selectedRoom = &ChatRoom{}
					*m.selectedRoom = selectedItem.(ChatRoom)
					m.state = "chat"
					m.loadMessages()
					m.updateViewport()
					return m, nil
				}
			}
		case "chat":
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = "list"
				m.selectedRoom = nil
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		if m.state == "list" {
			h, v := docStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4 // Adjust for header and footer
		}
		m.updateViewport()
	}

	var cmd tea.Cmd
	if m.state == "list" {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m *model) loadMessages() {
	m.messages = []Message{}
	rows, err := m.db.Query("SELECT id, chat_room_id, message, created_at FROM messages WHERE chat_room_id = ? ORDER BY created_at", m.selectedRoom.ID)
	if err != nil {
		log.Printf("Error loading messages: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ChatRoomID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		m.messages = append(m.messages, msg)
	}
}

// * ここでmessagesの表示をしている
func (m *model) updateViewport() {
	var content strings.Builder
	for _, msg := range m.messages {
		formattedTime := msg.CreatedAt.Format("2006-01-02 15:04:05")
		header := fmt.Sprintf("**[%s]**\n\n", formattedTime)
		renderedContent := render.RenderMarkdown(header + msg.Content)
		content.WriteString(renderedContent)
	}
	m.viewport.SetContent(content.String())
}

func (m model) headerView() string {
	title := titleStyle.Render("Chat Room: " + m.selectedRoom.Name)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m model) View() string {
	switch m.state {
	case "list":
		return docStyle.Render(m.list.View())
	case "chat":
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	default:
		return "Loading..."
	}
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

	m := model{
		list:  list.New(items, list.NewDefaultDelegate(), 0, 0),
		state: "list",
		db:    DbConnection,
	}
	m.list.Title = "Chat Rooms"
	m.viewport = viewport.New(80, 20)
	m.viewport.SetContent("No messages yet.")

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

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
