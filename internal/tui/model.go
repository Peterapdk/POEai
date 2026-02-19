package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

type model struct {
	conn      *websocket.Conn
	viewport  viewport.Model
	textinput textinput.Model
	messages  []string
	err       error
}

type msgReceived struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewModel(conn *websocket.Conn) model {
	ti := textinput.New()
	ti.Placeholder = "Say something to Poe..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to the Raven Hotel. Poe is at your service.\n")

	return model{
		conn:      conn,
		viewport:  vp,
		textinput: ti,
		messages:  []string{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.waitForMessage())
}

func (m model) waitForMessage() tea.Cmd {
	return func() tea.Msg {
		var msg msgReceived
		if err := m.conn.ReadJSON(&msg); err != nil {
			return err
		}
		return msg
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textinput, tiCmd = m.textinput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			content := m.textinput.Value()
			if content == "" {
				break
			}
			m.messages = append(m.messages, styleUserMsg.Render("You: ")+content)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Send to gateway
			m.conn.WriteJSON(msgReceived{Role: "user", Content: content})

			m.textinput.Reset()
		}

	case msgReceived:
		m.messages = append(m.messages, stylePoeMsg.Render("Poe: ")+msg.Content)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, m.waitForMessage()

	case error:
		m.err = msg
		return m, tea.Quit
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n%s",
		styleHeader.Render("POE — RAVEN HOTEL"),
		m.viewport.View(),
		m.textinput.View(),
		styleStatusBar.Render("(ctrl+c to quit)"),
	)
}

func Run(conn *websocket.Conn) error {
	p := tea.NewProgram(NewModel(conn), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
