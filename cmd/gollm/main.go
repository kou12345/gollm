package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/kou12345/gollm/pkg/utils"

	_ "github.com/mattn/go-sqlite3"
)

// 複雑なANSIエスケープシーケンスを処理する場合を除き、
// 通常はこれを使用する必要はありません。
// ちらつきに気づいた場合は有効にしてください。
//
// また、高性能レンダリングは端末の全サイズを使用するプログラムでのみ
// 機能することに注意してください。以下でtea.EnterAltScreen()を使用して
// これを有効にしています。
const useHighPerformanceRenderer = false

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

	docStyle = lipgloss.NewStyle().Margin(1, 2)
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

type State string

const (
	StateList State = "list"
	StateChat State = "chat"
)

type model struct {
	ready     bool           // ビューポートが初期化されたかどうか
	viewport  viewport.Model // ビューポートは、スクロール可能なビューを提供します
	chatRooms list.Model     // チャットルームのリスト
	messages  []Message      // チャットメッセージのリスト
	state     State          // アプリケーションの状態
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// このプログラムはビューポートの全サイズを使用しているため、
			// ビューポートを初期化する前にウィンドウの寸法を受け取る必要があります。
			// 初期寸法は非同期ですが素早く到着するため、ここで待機しています。
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer

			m.ready = true

			// m.listの要素を表示する

			// これは高性能レンダリングにのみ必要で、
			// ほとんどの場合は必要ありません。
			//
			// ビューポートをヘッダーの1行下にレンダリングします。
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// ビューポート全体をレンダリング（または再レンダリング）します。
			// ビューポートの初期化時とウィンドウのサイズ変更時の両方で必要です。
			//
			// これは高性能レンダリングにのみ必要です。
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	if m.state == StateList {
		m.chatRooms, cmd = m.chatRooms.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  初期化中..."
	}

	switch m.state {
	case StateList:
		return docStyle.Render(m.chatRooms.View())
	case StateChat:
		return "hoge"
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// ダミーメッセージを生成する関数
func generateDummyMessages() []Message {
	return []Message{
		{ID: 1, ChatRoomID: 1, Content: "こんにちは！", CreatedAt: time.Now().Add(-10 * time.Minute)},
		{ID: 2, ChatRoomID: 1, Content: "今日はどうですか？", CreatedAt: time.Now().Add(-5 * time.Minute)},
		{ID: 3, ChatRoomID: 1, Content: "素晴らしいです！", CreatedAt: time.Now().Add(-1 * time.Minute)},
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(utils.ErrorColor("Error loading .env file"))
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal(utils.ErrorColor("GEMINI_API_KEY is not set in the environment"))
	}

	DbConnection, err := sql.Open("sqlite3", "./db.sql")
	if err != nil {
		log.Fatal(utils.ErrorColor("Error opening database: " + err.Error()))
	}
	defer DbConnection.Close()

	cmd := `SELECT * FROM chat_rooms;`
	rows, err := DbConnection.Query(cmd)
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	fmt.Println(rows)

	var items []list.Item
	for rows.Next() {
		var chatRoom ChatRoom
		err = rows.Scan(&chatRoom.ID, &chatRoom.Name, &chatRoom.CreatedAt)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, chatRoom)
	}

	// ダミーメッセージを生成
	dummyMessages := generateDummyMessages()

	m := model{
		chatRooms: list.New(items, list.NewDefaultDelegate(), 0, 0),
		messages:  dummyMessages,
		state:     "list",
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // 端末の「代替画面バッファ」のフルサイズを使用します
		tea.WithMouseCellMotion(), // マウスホイールを追跡できるようにマウスサポートをオンにします
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("プログラムを実行できませんでした:", err)
		os.Exit(1)
	}
}
