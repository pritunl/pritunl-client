package iface

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

// LogsMsg carries a profile log fetch result.
type LogsMsg struct {
	Id   string
	Data string
	Err  error
}

// LogsCloseMsg is emitted when the log view is dismissed.
type LogsCloseMsg struct{}

// LogsClearMsg requests confirmation before clearing the profile log.
type LogsClearMsg struct {
	Sprofile *sprofile.Sprofile
}

type LogsView struct {
	sprfl    *sprofile.Sprofile
	viewport viewport.Model
	data     string
	follow   bool
	loading  bool
	width    int
	height   int
}

func NewLogsView(sprfl *sprofile.Sprofile, width, height int) LogsView {
	l := LogsView{
		sprfl:   sprfl,
		follow:  true,
		loading: true,
	}
	l.viewport = viewport.New(width, max(height-2, 1))
	l.viewport.SetContent("Loading...")
	l.SetSize(width, height)
	return l
}

func (l *LogsView) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.viewport.Width = width
	l.viewport.Height = max(height-2, 1)
	if l.follow {
		l.viewport.GotoBottom()
	}
}

func (l *LogsView) SetData(data string) {
	if data == l.data && !l.loading {
		return
	}
	l.loading = false
	l.data = data

	content := strings.TrimRight(data, "\n")
	if content == "" {
		content = "No log output"
	}

	// Wrap long lines so the page never scrolls horizontally.
	content = lipgloss.NewStyle().Width(l.width).Render(content)

	l.viewport.SetContent(content)
	if l.follow {
		l.viewport.GotoBottom()
	}
}

func fetchLogsCmd(sprfl *sprofile.Sprofile) tea.Cmd {
	return func() tea.Msg {
		data, err := sprfl.GetLogs()
		return LogsMsg{
			Id:   sprfl.Id,
			Data: data,
			Err:  err,
		}
	}
}

func clearLogsCmd(sprfl *sprofile.Sprofile) tea.Cmd {
	return func() tea.Msg {
		err := sprfl.ClearLogs()
		if err != nil {
			return ActionDoneMsg{
				Action: "Clear logs",
				Err:    err,
			}
		}
		return LogsMsg{
			Id:   sprfl.Id,
			Data: "",
		}
	}
}

func (l LogsView) Update(msg tea.Msg) (LogsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, logsKeys.Quit):
			return l, tea.Quit
		case key.Matches(msg, logsKeys.Back):
			return l, func() tea.Msg {
				return LogsCloseMsg{}
			}
		case key.Matches(msg, logsKeys.Clear):
			return l, func() tea.Msg {
				return LogsClearMsg{
					Sprofile: l.sprfl,
				}
			}
		case key.Matches(msg, logsKeys.Top):
			l.follow = false
			l.viewport.GotoTop()
			return l, nil
		case key.Matches(msg, logsKeys.End):
			l.follow = true
			l.viewport.GotoBottom()
			return l, nil
		}
	case LogsMsg:
		if msg.Id != l.sprfl.Id {
			return l, nil
		}
		if msg.Err == nil {
			l.SetData(msg.Data)
		}
		return l, nil
	}

	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	l.follow = l.viewport.AtBottom()

	return l, cmd
}

func (l LogsView) View() string {
	title := menuBarStyle.Width(l.width).Render(
		" Pritunl Client - Logs: " + l.sprfl.FormatedName())

	menu := renderMenuBar(l.width, []MenuItem{
		{Title: "Back", Key: "esc"},
		{Title: "Scroll", Key: "↑↓"},
		{Title: "Page", Key: "pgup/pgdn"},
		{Title: "Top", Key: "home"},
		{Title: "End", Key: "end"},
		{Title: "Clear", Key: "c"},
		{Title: "Quit", Key: "ctrl+c"},
	})

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		l.viewport.View(),
		menu,
	)
}
