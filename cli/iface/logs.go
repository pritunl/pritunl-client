package iface

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pritunl/pritunl-client/cli/logs"
	"github.com/pritunl/pritunl-client/cli/sprofile"
)

const (
	LogsService = "service"
	LogsClient  = "client"
)

// LogsMsg carries a log fetch result, Id is the profile id or one of
// LogsService and LogsClient.
type LogsMsg struct {
	Id   string
	Data string
	Err  error
}

// LogsCloseMsg is emitted when the log view is dismissed.
type LogsCloseMsg struct{}

// LogsClearMsg requests confirmation before clearing the current log.
type LogsClearMsg struct {
	Source LogsSource
}

// LogsSource is a selectable log in the log viewer mirroring the desktop
// client log viewer sources of service, client and each profile.
type LogsSource struct {
	Id    string
	Name  string
	Sprfl *sprofile.Sprofile
}

func (s LogsSource) Get() (string, error) {
	switch s.Id {
	case LogsService:
		return logs.GetServiceLog()
	case LogsClient:
		return logs.GetClientLog()
	default:
		return s.Sprfl.GetLogs()
	}
}

func (s LogsSource) Clear() error {
	switch s.Id {
	case LogsService:
		return logs.ClearServiceLog()
	case LogsClient:
		return logs.ClearClientLog()
	default:
		return s.Sprfl.ClearLogs()
	}
}

// LogsSources builds the log viewer sources from the current profiles.
func LogsSources(sprfls []*sprofile.Sprofile) []LogsSource {
	sources := []LogsSource{
		{Id: LogsService, Name: "Service"},
		{Id: LogsClient, Name: "Client"},
	}
	for _, sprfl := range sprfls {
		sources = append(sources, LogsSource{
			Id:    sprfl.Id,
			Name:  sprfl.FormatedName(),
			Sprfl: sprfl,
		})
	}
	return sources
}

type LogsView struct {
	sources  []LogsSource
	index    int
	viewport viewport.Model
	data     string
	follow   bool
	loading  bool
	width    int
	height   int
}

func NewLogsView(sources []LogsSource, id string,
	width, height int) LogsView {

	l := LogsView{
		sources: sources,
		follow:  true,
		loading: true,
	}
	l.viewport = viewport.New(width, max(height-2, 1))
	l.viewport.SetContent("Loading...")
	l.SetSize(width, height)
	l.setSource(id)
	return l
}

func (l *LogsView) Source() LogsSource {
	if l.index < 0 || l.index >= len(l.sources) {
		return LogsSource{Id: LogsService, Name: "Service"}
	}
	return l.sources[l.index]
}

func (l *LogsView) setSource(id string) {
	for i, src := range l.sources {
		if src.Id == id {
			l.index = i
			return
		}
	}
	l.index = 0
}

// SetSources refreshes the profile sources after a sync keeping the
// current selection, falling back to the service log when the profile is
// removed.
func (l *LogsView) SetSources(sources []LogsSource) {
	id := l.Source().Id
	l.sources = sources
	l.setSource(id)
}

func (l *LogsView) cycle(dir int) {
	if len(l.sources) == 0 {
		return
	}
	n := len(l.sources)
	l.index = ((l.index+dir)%n + n) % n
	l.loading = true
	l.follow = true
	l.viewport.SetContent("Loading...")
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

func fetchLogsCmd(src LogsSource) tea.Cmd {
	return func() tea.Msg {
		data, err := src.Get()
		return LogsMsg{
			Id:   src.Id,
			Data: data,
			Err:  err,
		}
	}
}

func clearLogsCmd(src LogsSource) tea.Cmd {
	return func() tea.Msg {
		err := src.Clear()
		if err != nil {
			return ActionDoneMsg{
				Action: "Clear logs",
				Err:    err,
			}
		}
		return LogsMsg{
			Id:   src.Id,
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
			src := l.Source()
			return l, func() tea.Msg {
				return LogsClearMsg{
					Source: src,
				}
			}
		case key.Matches(msg, logsKeys.Next):
			l.cycle(1)
			return l, fetchLogsCmd(l.Source())
		case key.Matches(msg, logsKeys.Prev):
			l.cycle(-1)
			return l, fetchLogsCmd(l.Source())
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
		if msg.Id != l.Source().Id {
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
		" Pritunl Client - Logs: " + l.Source().Name)

	menu := renderMenuBar(l.width, []MenuItem{
		{Title: "Back", Key: "esc"},
		{Title: "Change Log", Key: "←/→"},
		{Title: "Scroll", Key: "↑/↓"},
		{Title: "Page", Key: "pgup/pgdn"},
		{Title: "Top", Key: "home"},
		{Title: "End", Key: "end"},
		{Title: "Clear", Key: "c"},
	})

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		l.viewport.View(),
		menu,
	)
}
